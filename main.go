package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("GTA Website Converter")
		fmt.Println("Usage:")
		fmt.Println("  whm2html iv   - Convert GTA IV websites (.whm files)")
		fmt.Println("  whm2html v    - Convert GTA V websites (.gfx files)")
		fmt.Println("  whm2html both - Convert both GTA IV and V websites")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "iv":
		if err := ProcessGTAIV(); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing GTA IV: %v\n", err)
			os.Exit(1)
		}
	case "v":
		if err := ProcessGTAV(); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing GTA V: %v\n", err)
			os.Exit(1)
		}
	case "both":
		fmt.Println("Processing GTA IV...")
		if err := ProcessGTAIV(); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing GTA IV: %v\n", err)
		}
		fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
		fmt.Println("Processing GTA V...")
		if err := ProcessGTAV(); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing GTA V: %v\n", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Println("Valid commands: iv, v, both")
		os.Exit(1)
	}
}
