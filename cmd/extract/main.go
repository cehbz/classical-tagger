package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cehbz/classical-tagger/internal/scraping"
)

var (
	url        = flag.String("url", "", "URL to extract from (use this OR -dir)")
	dir        = flag.String("dir", "", "Local directory with FLAC files to extract from (use this OR -url)")
	outputFile = flag.String("output", "", "Output JSON file (default: stdout)")
	validate   = flag.Bool("validate", true, "Validate extracted metadata against domain rules")
	verbose    = flag.Bool("verbose", false, "Verbose output including parsing notes")
	force      = flag.Bool("force", false, "Create output even with required field errors")
	timeout    = flag.Duration("timeout", 30*time.Second, "HTTP request timeout (URL mode only)")
)

func main() {
	flag.Parse()

	// Validate input: need either -url or -dir
	if *url == "" && *dir == "" {
		printUsage()
		os.Exit(1)
	}

	if *url != "" && *dir != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both -url and -dir flags\n")
		os.Exit(1)
	}

	// Route to appropriate extraction method
	if *dir != "" {
		extractFromDirectory()
	} else {
		extractFromURL()
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Error: Either -url or -dir flag is required\n")
	fmt.Fprintf(os.Stderr, "\nUsage:\n")
	fmt.Fprintf(os.Stderr, "  extract -url URL [options]     # Extract from website\n")
	fmt.Fprintf(os.Stderr, "  extract -dir PATH [options]    # Extract from local FLAC files\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nSupported sites:\n")
	fmt.Fprintf(os.Stderr, "  - Harmonia Mundi (harmoniamundi.com)\n")
	fmt.Fprintf(os.Stderr, "  - Presto Classical (prestoclassical.co.uk)\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Extract from website:\n")
	fmt.Fprintf(os.Stderr, "  extract -url \"https://prestomusic.com/...\" -output album.json\n\n")
	fmt.Fprintf(os.Stderr, "  # Extract from existing FLAC directory:\n")
	fmt.Fprintf(os.Stderr, "  extract -dir \"/music/Bach - Goldberg Variations\" -output album.json\n")
}

// Parser interface for HTML parsing
type Parser interface {
	Parse(html string) (*scraping.ExtractionResult, error)
}

// extractFromDirectory extracts metadata from local FLAC files
func extractFromDirectory() {
	fmt.Printf("Extracting from local directory: %s\n", *dir)

	// Create local extractor
	extractor := scraping.NewLocalExtractor()

	// Extract metadata
	result, err := extractor.ExtractFromDirectory(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting from directory: %v\n", err)
		os.Exit(1)
	}

	// Process and output the result
	processResult(result, "local directory")
}

// extractFromURL extracts metadata from a website
func extractFromURL() {
	// Determine which parser to use based on URL
	if *verbose {
		fmt.Printf("Analyzing URL: %s\n", *url)
	}

	var parser Parser
	var siteName string

	if isHarmoniaMundi(*url) {
		parser = scraping.NewHarmoniaMundiParser()
		siteName = "Harmonia Mundi"
	} else if isPrestoClassical(*url) {
		parser = scraping.NewPrestoClassicalParser()
		siteName = "Presto Classical"
	} else {
		fmt.Fprintf(os.Stderr, "Error: No parser available for this URL\n")
		fmt.Fprintf(os.Stderr, "Supported sites:\n")
		fmt.Fprintf(os.Stderr, "  - harmoniamundi.com\n")
		fmt.Fprintf(os.Stderr, "  - prestoclassical.co.uk\n")
		fmt.Fprintf(os.Stderr, "\nTo add support for more sites, see METADATA_SOURCES.md\n")
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Using parser: %s\n\n", siteName)
	}

	// Fetch HTML
	fmt.Printf("Fetching %s...\n", siteName)
	html, err := fetchHTML(*url, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("✓ Downloaded %d bytes\n\n", len(html))
	}

	// Parse HTML
	fmt.Printf("Parsing metadata...\n")
	result, err := parser.Parse(html)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML: %v\n", err)
		os.Exit(1)
	}

	// Process and output the result
	processResult(result, siteName)
}

// processResult handles the common result processing and output
func processResult(result *scraping.ExtractionResult, source string) {
	data := result.Data()
	
	// Display extraction summary
	fmt.Printf("✓ Extracted: %s", data.Title)
	if data.OriginalYear > 0 {
		fmt.Printf(" (%d)", data.OriginalYear)
	}
	fmt.Println()

	if data.Edition != nil {
		if data.Edition.Label != "" {
			fmt.Printf("  Label: %s", data.Edition.Label)
			if data.Edition.CatalogNumber != "" {
				fmt.Printf(" %s", data.Edition.CatalogNumber)
			}
			fmt.Println()
		} else if data.Edition.CatalogNumber != "" {
			fmt.Printf("  Catalog: %s\n", data.Edition.CatalogNumber)
		}
	}
	fmt.Printf("  Tracks: %d\n", len(data.Tracks))

	// Show extraction errors and warnings
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

	for _, w := range result.Warnings() {
		fmt.Printf("  WARNING: %s\n", w)
	}

	// Fail if required errors and not forced
	if result.HasRequiredErrors() && !*force {
		fmt.Fprintf(os.Stderr, "\n❌ Extraction failed due to required field errors.\n")
		fmt.Fprintf(os.Stderr, "Use -force to create output file anyway (not recommended).\n")
		os.Exit(1)
	}

	// Show parsing notes in verbose mode
	if *verbose && result.ParsingNotes() != nil {
		fmt.Println("\nParsing Notes:")
		if jsonBytes, err := json.MarshalIndent(result.ParsingNotes(), "  ", "  "); err == nil {
			fmt.Printf("  %s\n", string(jsonBytes))
		}
	}

	// Validate against domain model if requested
	if *validate {
		fmt.Println("\nValidating against domain model...")
		album, err := data.ToAlbum()
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Domain conversion failed: %v\n", err)
			if !*force {
				fmt.Fprintf(os.Stderr, "Use -force to create output anyway, or fix the errors above.\n")
				os.Exit(1)
			}
		} else {
			// Run validation
			issues := album.Validate()
			if len(issues) == 0 {
				fmt.Println("✓ Metadata is valid")
			} else {
				errorCount := 0
				warningCount := 0
				for _, issue := range issues {
					if issue.Level().String() == "ERROR" {
						errorCount++
					} else if issue.Level().String() == "WARNING" {
						warningCount++
					}
				}

				if errorCount > 0 {
					fmt.Printf("⚠️  Found %d errors and %d warnings\n", errorCount, warningCount)
					if *verbose {
						for _, issue := range issues {
							fmt.Printf("  %s\n", issue.String())
						}
					}
				} else {
					fmt.Printf("✓ Valid (with %d warnings)\n", warningCount)
				}
			}
		}
	}

	// Convert to JSON
	fmt.Println("\nConverting to JSON...")
	jsonData, err := scraping.SaveToJSON(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

	// Output
	if *outputFile == "" {
		// Write to stdout
		fmt.Println("\n" + string(jsonData))
	} else {
		// Write to file
		if err := os.WriteFile(*outputFile, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}

		absPath, _ := filepath.Abs(*outputFile)
		fmt.Printf("✓ Saved to: %s\n", absPath)

		// Show next steps
		fmt.Println("\n📋 Next steps:")
		fmt.Printf("  1. Review and edit: %s\n", *outputFile)
		fmt.Printf("  2. Validate album directory: validate /path/to/album\n")
		fmt.Printf("  3. Apply tags: tag -metadata %s -dir /path/to/album\n", *outputFile)
	}
}

// fetchHTML fetches HTML content from a URL
func fetchHTML(url string, timeout time.Duration) (string, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set a reasonable User-Agent
	req.Header.Set("User-Agent", "classical-tagger/0.1 (metadata extraction tool)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// isHarmoniaMundi checks if URL is from Harmonia Mundi
func isHarmoniaMundi(url string) bool {
	return contains(url, "harmoniamundi.com")
}

// isPrestoClassical checks if URL is from Presto Classical
func isPrestoClassical(url string) bool {
	return contains(url, "prestoclassical.co.uk") || contains(url, "prestomusic.com")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}