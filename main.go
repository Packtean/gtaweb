package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	sourceDir := "iv/html"
	targetDir := "iv"

	fmt.Print("Loading localization strings... ")
	if err := LoadLocalizationFile("iv/american.txt"); err != nil {
		fmt.Printf("Warning: %v\n", err)
	} else {
		fmt.Printf("Loaded %d strings\n", len(localizationMap))
	}

	fmt.Printf("Source: %s\n", sourceDir)
	fmt.Printf("Target: %s\n\n", targetDir)

	// walk through all .whm files
	numErrors := 0
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".whm") {
			return nil
		}

		// skip non-www directories
		relPath, _ := filepath.Rel(sourceDir, path)
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(parts) < 1 || !strings.HasPrefix(parts[0], "www.") {
			return nil
		}

		depth := len(parts)

		fmt.Printf("Processing: %s\n", path)

		// read and parse the WHM file
		doc, err := ReadWHMFile(path)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			numErrors++
			return nil
		}

		// determine output path, e.g.
		// iv/html/www.example.com/index.whm -> iv/www.example.com/index.html
		outputPath := filepath.Join(targetDir, relPath)
		outputPath = strings.TrimSuffix(outputPath, ".whm") + ".html"

		// create output directory
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("  ERROR creating directory: %v\n", err)
			numErrors++
			return nil
		}

		// export to HTML
		if err := ExportHTML(doc, outputPath, depth); err != nil {
			fmt.Printf("  ERROR exporting: %v\n", err)
			numErrors++
			return nil
		}

		fmt.Printf("  -> %s\n", outputPath)

		// export textures if present
		if doc.TextureDictionary != nil && len(doc.TextureDictionary.Textures) > 0 {
			textureDir := filepath.Dir(outputPath)
			textureCount := 0

			for _, tex := range doc.TextureDictionary.Textures {
				if len(tex.TextureData) == 0 {
					fmt.Printf("  ERROR: Texture %s is empty\n", tex.Name)
					numErrors++
					continue
				}
				// the texture name may include path components, e.g. Image/filename
				texPath := filepath.Join(textureDir, tex.Name+".png")
				texDir := filepath.Dir(texPath)
				os.MkdirAll(texDir, 0755)
				if err := tex.DecodeToPNG(texPath); err == nil {
					textureCount++
				}
			}
			if textureCount > 0 {
				fmt.Printf("  -> %d textures\n", textureCount)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if numErrors > 0 {
		fmt.Printf("\nConversion complete with %d errors!\n", numErrors)
	} else {
		fmt.Println("\nConversion complete.")
	}

	fmt.Println("\nGenerating index pages...")
	if err := GenerateIndexes(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating indexes: %v\n", err)
		os.Exit(1)
	}
}
