package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cehbz/classical-tagger/internal/scraping"
)

var (
	url        = flag.String("url", "", "URL to extract from")
	outputFile = flag.String("output", "", "Output JSON file")
	validate   = flag.Bool("validate", true, "Validate extracted metadata")
	verbose    = flag.Bool("verbose", false, "Verbose output")
	force      = flag.Bool("force", false, "Create output even with errors")
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
	result, err := extractor.Extract(*url) // ← Now returns ExtractionResult
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Show extraction results
	data := result.Data()
	fmt.Printf("✓ Extracted: %s (%d)\n", data.Title, data.OriginalYear)
	fmt.Printf("  Tracks: %d\n", len(data.Tracks))

	// Show errors
	if result.HasErrors() {
		fmt.Println("\n⚠ Extraction Issues:")
		for _, e := range result.Errors() {
			if e.Required() {
				fmt.Printf("  ERROR: %s - %s\n", e.Field(), e.Message())
			} else {
				fmt.Printf("  WARNING: %s - %s\n", e.Field(), e.Message())
			}
		}
	}

	// Show warnings
	for _, w := range result.Warnings() {
		fmt.Printf("  WARNING: %s\n", w)
	}

	// Fail if required errors and not forced
	if result.HasRequiredErrors() && !*force {
		fmt.Fprintf(os.Stderr, "\nExtraction failed due to required field errors.\n")
		fmt.Fprintf(os.Stderr, "Use -force to create output file anyway.\n")
		os.Exit(1)
	}

	// Show parsing notes in verbose mode
	if *verbose && result.ParsingNotes() != nil {
		fmt.Println("\nParsing Notes:")
		// Pretty print the notes
		if jsonBytes, err := json.MarshalIndent(result.ParsingNotes(), "  ", "  "); err == nil {
			fmt.Println(string(jsonBytes))
		}
	}

	// Convert to domain album (may fail validation)
	if _, err = data.ToAlbum(); err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to domain: %v\n", err)
		if !*force {
			os.Exit(1)
		}
	}

	// Save to JSON
	fmt.Println("Converting to JSON...")
	jsonData, err := scraping.SaveToJSON(data)
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
