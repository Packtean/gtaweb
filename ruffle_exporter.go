package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ruffleHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body, html {
            background-color: #000;
            overflow: hidden;
            width: 100%;
            height: 100%;
        }

        #container {
            width: 100vw;
            height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
        }

        #ruffle-container {
            width: 100%;
            height: 100%;
        }

        #ruffle-embed {
            width: 100%;
            height: 100%;
            display: block;
        }

        #loading {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            color: #fff;
            font-size: 18px;
            text-align: center;
            z-index: 10;
        }
    </style>
    <script src="{{.RufflePath}}"></script>
</head>
<body>
    <div id="container">
        <div id="loading">
            <div class="spinner"></div>
            <p>Loading {{.Title}}...</p>
        </div>
        <div id="ruffle-container">
            <embed
                id="ruffle-embed"
                src="{{.SWFPath}}"
                width="100%"
                height="100%"
                type="application/x-shockwave-flash"
                allowfullscreen="true"
                quality="high"
                bgcolor="#000000"
                wmode="direct"
            />
        </div>
    </div>

    <script>
        window.RufflePlayer = window.RufflePlayer || {};
        window.RufflePlayer.config = {
            autoplay: "on",
            unmuteOverlay: "hidden",
            logLevel: "error",
            letterbox: "on",
            scale: "showAll",
            warnOnUnsupportedContent: false,
            contextMenu: "off",
            base: "/v/"
        };

        window.addEventListener('load', function() {
            setTimeout(function() {
                const loader = document.getElementById('loading');
                if (loader) loader.style.display = 'none';
            }, 1000);
        });
    </script>
</body>
</html>
`

// RuffleExportOptions contains options for exporting with Ruffle
type RuffleExportOptions struct {
	Title      string
	SWFPath    string
	RufflePath string
}

// ExportGFX exports a GFX file as an HTML page
func ExportGFX(gfxPath, outputPath string, depth int) error {
	gfx, err := ReadGFXFile(gfxPath)
	if err != nil {
		return fmt.Errorf("failed to read GFX file: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(gfxPath), ".gfx")
	swfPath := filepath.Join(outputDir, baseName+".swf")

	if err := gfx.ConvertToSWF(swfPath); err != nil {
		return fmt.Errorf("failed to convert to SWF: %w", err)
	}

	relativePrefix := strings.Repeat("../", depth)
	rufflePath := relativePrefix + "ruffle/ruffle.js"
	relativeSWFPath := baseName + ".swf"

	title := strings.ReplaceAll(baseName, "_", ".")
	title = strings.ReplaceAll(title, "www.", "")

	options := RuffleExportOptions{
		Title:      title,
		SWFPath:    relativeSWFPath,
		RufflePath: rufflePath,
	}

	htmlContent := expandTemplate(ruffleHTMLTemplate, options)

	if err := os.WriteFile(outputPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

func expandTemplate(template string, opts RuffleExportOptions) string {
	result := template
	result = strings.ReplaceAll(result, "{{.Title}}", opts.Title)
	result = strings.ReplaceAll(result, "{{.SWFPath}}", opts.SWFPath)
	result = strings.ReplaceAll(result, "{{.RufflePath}}", opts.RufflePath)
	return result
}
