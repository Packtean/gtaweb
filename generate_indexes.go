package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenerateIndexes generates root, IV or V index.html pages
func GenerateIndexes(iv bool) error {
	if err := generateRootIndex(); err != nil {
		return fmt.Errorf("error generating root index: %w", err)
	}
	if err := generateIndex(iv); err != nil {
		return fmt.Errorf("error generating IV index: %w", err)
	}

	fmt.Println("Index pages generated successfully!")
	return nil
}

func generateRootIndex() error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GTA Web Archive</title>
    <style>
        @font-face {
            font-family: 'DIN Medium';
            font-style: normal;
            font-weight: normal;
            src: local('DIN Medium'), url('ttf/DIN-Medium.woff2') format('woff2');
        }
        body {
            font-family: 'Din Medium', Arial, Helvetica, sans-serif;
            background-color: #f5f5f5;
            margin: 40px;
            line-height: 1.6;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            border: 1px solid #ddd;
        }
        h1 {
            font-size: 32px;
            color: #333;
            margin-bottom: 10px;
            border-bottom: 2px solid #333;
            padding-bottom: 10px;
        }
        .subtitle {
            font-size: 16px;
            color: #666;
            margin-bottom: 30px;
        }
        .game-list {
            margin: 20px 0;
        }
        .game-card {
            display: block;
            background: #fff;
            border: 1px solid #ccc;
            padding: 20px;
            margin-bottom: 15px;
            text-decoration: none;
            color: #333;
        }
        .game-card:hover {
            background: #f9f9f9;
            border-color: #999;
        }
        .game-title {
            font-size: 20px;
            font-weight: bold;
            margin-bottom: 5px;
            color: #0066cc;
        }
        .game-description {
            font-size: 14px;
            color: #666;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            font-size: 12px;
            color: #999;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>GTA Web Archive</h1>
        <p class="subtitle">Explore the in-game internet from Grand Theft Auto</p>

        <div class="game-list">
            <a href="iv/index.html" class="game-card">
                <div class="game-title">Grand Theft Auto IV</div>
                <div class="game-description">Browse the websites from Liberty City</div>
            </a>
            <a href="v/index.html" class="game-card">
                <div class="game-title">Grand Theft Auto V</div>
                <div class="game-description">Browse the websites from Los Santos (WIP)</div>
            </a>
        </div>

        <div class="footer">
            Note: your adblocker may block the ingame ads, make sure to disable it.
        </div>
    </div>
</body>
</html>
`
	return os.WriteFile("index.html", []byte(html), 0644)
}

func generateIndex(iv bool) error {
	dirName := "iv"
	if !iv {
		dirName = "v"
	}

	var sites []string
	err := filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(dirName, path)
		if strings.HasPrefix(d.Name(), "www.") && !strings.Contains(relPath, string(filepath.Separator)) {
			sites = append(sites, d.Name())
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directories: %w", err)
	}

	sort.Strings(sites)

	cityName := "Libery City"
	if !iv {
		cityName = "Los Santos"
	}
	game := strings.ToUpper(dirName)

	var html strings.Builder
	fmt.Fprintf(&html, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GTA %s - Website Directory</title>
    <style>
        @font-face {
            font-family: 'DIN Medium';
            font-style: normal;
            font-weight: normal;
            src: local('DIN Medium'), url('../ttf/DIN-Medium.woff2') format('woff2');
        }
        body {
            font-family: 'Din Medium', Arial, Helvetica, sans-serif;
            background-color: #f5f5f5;
            margin: 40px;
            line-height: 1.6;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            border: 1px solid #ddd;
        }
        h1 {
            font-size: 32px;
            color: #333;
            margin-bottom: 10px;
            border-bottom: 2px solid #333;
            padding-bottom: 10px;
        }
        .subtitle {
            font-size: 16px;
            color: #666;
            margin-bottom: 10px;
        }
        .back-link {
            display: inline-block;
            color: #0066cc;
            text-decoration: none;
            font-size: 14px;
            margin-bottom: 20px;
        }
        .back-link:hover {
            text-decoration: underline;
        }
        .count {
            display: inline-block;
            background: #eee;
            padding: 5px 10px;
            font-size: 14px;
            color: #666;
            margin-bottom: 20px;
        }
        .site-list {
            list-style: none;
            padding: 0;
            margin: 20px 0;
            column-count: 3;
            column-gap: 20px;
        }
        .site-list li {
            break-inside: avoid;
            margin-bottom: 5px;
        }
        .site-link {
            display: block;
            padding: 8px 10px;
            text-decoration: none;
            color: #0066cc;
            background: #f9f9f9;
            border: 1px solid #e0e0e0;
        }
        .site-link:hover {
            background: #fff;
            border-color: #0066cc;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            font-size: 12px;
            color: #999;
            text-align: center;
        }
        @media (max-width: 900px) {
            .site-list {
                column-count: 2;
            }
        }
        @media (max-width: 600px) {
            .site-list {
                column-count: 1;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>GTA %s Websites</h1>
        <p class="subtitle">%s's Internet Archive</p>
        <div class="count">`, game, game, cityName)
	fmt.Fprintf(&html, "%d websites available", len(sites))
	html.WriteString(`</div>
        <br>
        <a href="../index.html" class="back-link">← Back to Game Selection</a>

        <ul class="site-list">
`)

	for _, site := range sites {
		// try to find index.html in the site directory
		indexPath := filepath.Join(site, "index.html")
		displayName := strings.TrimPrefix(site, "www.")

		fmt.Fprintf(&html, `            <li><a href="%s" class="site-link">%s</a></li>
`, indexPath, displayName)
	}

	fmt.Fprintf(&html, `        </ul>

        <div class="footer">
            These websites were extracted from GTA %s and converted to HTML format. Note: your adblocker may block the ingame ads, make sure to disable it.
        </div>
    </div>
</body>
</html>
`, game)

	return os.WriteFile(fmt.Sprintf("%s/index.html", dirName), []byte(html.String()), 0644)
}
