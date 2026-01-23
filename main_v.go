package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ProcessGTAV processes GTA V .gfx files
func ProcessGTAV() error {
	sourceDir := "v"
	targetDir := "v"

	fmt.Printf("Source: %s\n", sourceDir)
	fmt.Printf("Target: %s\n\n", targetDir)

	fmt.Print("Converting gfxfontlib.gfx... ")
	gfxFontLibPath := "v/gfxfontlib.gfx"
	if _, err := os.Stat(gfxFontLibPath); err == nil {
		gfxFontLib, err := ReadGFXFile(gfxFontLibPath)
		if err == nil {
			fontLibSWFPath := filepath.Join(targetDir, "gfxfontlib.swf")
			if err := gfxFontLib.ConvertToSWF(fontLibSWFPath); err != nil {
				fmt.Printf("Warning: %v\n", err)
			} else {
				fmt.Println("Done")
			}
		} else {
			fmt.Printf("Warning: %v\n", err)
		}
	} else {
		fmt.Println("Not found, skipping")
	}

	// walk through all .gfx files
	numProcessed := 0
	numErrors := 0
	numSkipped := 0

	directories := []string{
		"v/x64b_scaleform_web",
		"v/update_scaleform_web",
		"v/update_copy_scaleform_web",
	}

	for _, dir := range directories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("Skipping non-existent directory: %s\n", dir)
			continue
		}

		dirName := filepath.Base(dir)
		fmt.Printf("\n--- Processing %s\n", dirName)

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// process .ytd texture files
			if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".ytd") {
				baseName := strings.TrimSuffix(d.Name(), ".ytd")
				if strings.HasPrefix(strings.ToLower(baseName), "www_") ||
					strings.HasPrefix(strings.ToLower(baseName), "foreclosures_") {

					websiteName := strings.ReplaceAll(baseName, "_d_", "-")
					websiteName = strings.ReplaceAll(websiteName, "_", ".")
					outputDir := filepath.Join(targetDir, websiteName)

					fmt.Printf("  Extracting textures from: %s\n", d.Name())
					if err := ExtractYTDTextures(path, outputDir); err != nil {
						fmt.Printf("    Warning: %v\n", err)
					}
				}
				return nil
			}

			if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".gfx") {
				return nil
			}

			if strings.HasPrefix(strings.ToLower(d.Name()), "web_browser") {
				numSkipped++
				return nil
			}

			baseName := strings.TrimSuffix(d.Name(), ".gfx")
			if !strings.HasPrefix(strings.ToLower(baseName), "www_") &&
				!strings.HasPrefix(strings.ToLower(baseName), "foreclosures_") {
				numSkipped++
				return nil
			}

			fmt.Printf("Processing: %s\n", d.Name())

			websiteName := strings.ReplaceAll(baseName, "_d_", "-")
			websiteName = strings.ReplaceAll(websiteName, "_", ".")

			// v/x64b_scaleform_web/www_example_com.gfx -> v/www.example.com/index.html
			outputDir := filepath.Join(targetDir, websiteName)
			outputPath := filepath.Join(outputDir, "index.html")

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("  ERROR creating directory: %v\n", err)
				numErrors++
				return nil
			}

			// v/output/www.example.com/index.html -> ../ruffle/ruffle.js
			depth := 1
			if err := ExportGFX(path, outputPath, depth); err != nil {
				fmt.Printf("  ERROR: %v\n", err)
				numErrors++
				return nil
			}

			fmt.Printf("  -> %s\n", outputPath)
			numProcessed++

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", dir, err)
			continue
		}
	}

	if numErrors > 0 {
		fmt.Printf("\nConversion complete with %d errors!\n", numErrors)
	} else {
		fmt.Println("\nConversion complete.")
	}

	if numProcessed > 0 {
		fmt.Println("\nGenerating index pages...")
		if err := GenerateIndexes(false); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating indexes: %v\n", err)
		}
	}

	return nil
}
