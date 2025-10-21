package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/scraping"
)

var (
	url        = flag.String("url", "", "URL to extract metadata from (required)")
	outputFile = flag.String("output", "", "Output JSON file path (default: stdout)")
	validate   = flag.Bool("validate", true, "Validate extracted metadata")
	verbose    = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()

	if *url == "" {
		fmt.Fprintf(os.Stderr, "Error: -url flag is required\n")
		fmt.Fprintf(os.Stderr, "\nUsage: extract -url URL [options]\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported sites:\n")
		fmt.Fprintf(os.Stderr, "  - Harmonia Mundi (harmoniamundi.com)\n")
		fmt.Fprintf(os.Stderr, "  - More coming soon...\n")
		os.Exit(1)
	}

	// Create registry with available extractors
	registry := scraping.DefaultRegistry()

	// Add Harmonia Mundi extractor
	registry.Register(scraping.NewHarmoniaMundiExtractor())

	// Find appropriate extractor
	if *verbose {
		fmt.Printf("Finding extractor for: %s\n", *url)
	}

	extractor := registry.Get(*url)
	if extractor == nil {
		fmt.Fprintf(os.Stderr, "Error: No extractor available for this URL\n")
		fmt.Fprintf(os.Stderr, "Supported sites:\n")
		fmt.Fprintf(os.Stderr, "  - harmoniamundi.com\n")
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Using extractor: %s\n\n", extractor.Name())
	}

	// Extract metadata
	fmt.Printf("Extracting metadata from %s...\n", extractor.Name())
	albumData, err := extractor.Extract(*url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Extracted: %s (%d)\n", albumData.Title, albumData.OriginalYear)
	fmt.Printf("  Tracks: %d\n", len(albumData.Tracks))
	if albumData.Edition != nil {
		fmt.Printf("  Label: %s\n", albumData.Edition.Label)
		fmt.Printf("  Catalog: %s\n", albumData.Edition.CatalogNumber)
	}
	fmt.Println()

	// Validate if requested
	if *validate {
		fmt.Println("Validating extracted metadata...")

		// Convert to domain album for validation
		album, err := albumData.ToAlbum()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting to album: %v\n", err)
			os.Exit(1)
		}

		issues := album.Validate()

		hasErrors := false
		for _, issue := range issues {
			if issue.Level() == domain.LevelError {
				hasErrors = true
				fmt.Printf("❌ %s\n", issue)
			} else if issue.Level() == domain.LevelWarning {
				fmt.Printf("⚠️  %s\n", issue)
			} else if *verbose {
				fmt.Printf("ℹ️  %s\n", issue)
			}
		}

		if hasErrors {
			fmt.Fprintf(os.Stderr, "\n⚠️  Extracted metadata has validation errors\n")
			fmt.Fprintf(os.Stderr, "You may need to manually fix the JSON before using it\n")
		} else if len(issues) == 0 {
			fmt.Println("✓ Metadata is valid")
		} else {
			fmt.Println("⚠️  Metadata has warnings but is usable")
		}
	}

	// Save to JSON
	fmt.Println("Converting to JSON...")
	jsonData, err := scraping.SaveToJSON(albumData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving JSON: %v\n", err)
		os.Exit(1)
	}

	// Output
	if *outputFile == "" {
		// Write to stdout
		fmt.Println(string(jsonData))
	} else {
		// Write to file
		err := os.WriteFile(*outputFile, jsonData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}

		absPath, _ := filepath.Abs(*outputFile)
		fmt.Printf("✓ Saved to: %s\n", absPath)
		fmt.Println("\nNext steps:")
		fmt.Printf("  1. Review and edit: %s\n", *outputFile)
		fmt.Printf("  2. Validate album: validate /path/to/album\n")
		fmt.Printf("  3. Apply tags: tag -metadata %s -dir /path/to/album\n", *outputFile)
	}
}
