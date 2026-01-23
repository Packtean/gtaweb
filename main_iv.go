package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ProcessGTAIV processes GTA IV .whm files
func ProcessGTAIV() error {
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

	numErrors := 0
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".whm") {
			return nil
		}

		relPath, _ := filepath.Rel(sourceDir, path)
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(parts) < 1 || !strings.HasPrefix(parts[0], "www.") {
			return nil
		}

		depth := len(parts)

		fmt.Printf("Processing: %s\n", path)

		doc, err := ReadWHMFile(path)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			numErrors++
			return nil
		}

		// iv/html/www.example.com/index.whm -> iv/www.example.com/index.html
		outputPath := filepath.Join(targetDir, relPath)
		outputPath = strings.TrimSuffix(outputPath, ".whm") + ".html"

		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("  ERROR creating directory: %v\n", err)
			numErrors++
			return nil
		}

		if err := ExportHTML(doc, outputPath, depth); err != nil {
			fmt.Printf("  ERROR exporting: %v\n", err)
			numErrors++
			return nil
		}

		fmt.Printf("  -> %s\n", outputPath)

		if doc.TextureDictionary != nil && len(doc.TextureDictionary.Textures) > 0 {
			textureDir := filepath.Dir(outputPath)
			textureCount := 0

			for _, tex := range doc.TextureDictionary.Textures {
				if len(tex.TextureData) == 0 {
					fmt.Printf("  ERROR: Texture %s is empty\n", tex.Name)
					numErrors++
					continue
				}
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
		return err
	}

	if numErrors > 0 {
		fmt.Printf("\nConversion complete with %d errors!\n", numErrors)
	} else {
		fmt.Println("\nConversion complete.")
	}

	fmt.Println("\nGenerating index pages...")
	if err := GenerateIndexes(true); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating indexes: %v\n", err)
		return err
	}

	return nil
}
