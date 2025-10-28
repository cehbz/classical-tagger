package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/scraping"
)

var (
	url        = flag.String("url", "", "URL to extract metadata from")
	file       = flag.String("file", "", "Local HTML file to parse")
	dir        = flag.String("dir", "", "Local directory to extract metadata from")
	outputFile = flag.String("output", "", "Output file for extracted metadata, stdout by default")
	timeout    = flag.Duration("timeout", 30*time.Second, "HTTP request timeout")
	verbose    = flag.Bool("verbose", false, "Enable verbose output")
	force      = flag.Bool("force", false, "Create output even if required fields are missing")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	// Count number of inputs specified
	inputCount := 0
	if *url != "" {
		inputCount++
	}
	if *file != "" {
		inputCount++
	}
	if *dir != "" {
		inputCount++
	}

	// Validate input
	if inputCount == 0 {
		fmt.Fprintf(os.Stderr, "Error: Must specify one of -url, -file, or -dir\n\n")
		usage()
		os.Exit(1)
	}

	if inputCount > 1 {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify more than one input source\n\n")
		usage()
		os.Exit(1)
	}

	// Route to appropriate extraction method
	if *file != "" {
		extractFromFile()
	} else if *dir != "" {
		extractFromDirectory()
	} else {
		extractFromURL()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: extract [options]\n\n")
	fmt.Fprintf(os.Stderr, "Extract metadata from websites, local HTML files, or FLAC directories.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nSupported websites:\n")
	fmt.Fprintf(os.Stderr, "  - prestomusic.com / prestoclassical.co.uk\n")
	fmt.Fprintf(os.Stderr, "  - discogs.com (use -file for manual workflow)\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Extract from website:\n")
	fmt.Fprintf(os.Stderr, "  extract -url \"https://prestomusic.com/...\" -output presto.json\n\n")
	fmt.Fprintf(os.Stderr, "  # Extract from saved HTML file (for Discogs):\n")
	fmt.Fprintf(os.Stderr, "  extract -file discogs_release.html -output discogs.json\n\n")
	fmt.Fprintf(os.Stderr, "  # Extract from existing FLAC directory:\n")
	fmt.Fprintf(os.Stderr, "  extract -dir \"/music/Bach - Goldberg Variations\" -output directory.json\n")
	fmt.Fprintf(os.Stderr, "\nNote: Discogs blocks automated requests. Save the page in your browser,\n")
	fmt.Fprintf(os.Stderr, "      then use -file to parse the saved HTML.\n")
}

// Parser interface for HTML parsing
type Parser interface {
	Parse(html string) (*scraping.ExtractionResult, error)
}

// extractFromDirectory extracts metadata from local FLAC files
func extractFromDirectory() {
	fmt.Fprintf(os.Stderr, "Extracting from local directory: %s\n", *dir)

	// Create local extractor
	extractor := scraping.NewLocalExtractor()

	// Extract metadata
	result, err := extractor.ExtractFromDirectory(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting from directory: %v\n", err)
		os.Exit(1)
	}

	// Process and output the result
	processResult(result)
}

// extractFromFile parses a locally saved HTML file
func extractFromFile() {
	fmt.Printf("Parsing local HTML file: %s\n", *file)

	// Read HTML file
	htmlBytes, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	html := string(htmlBytes)

	if *verbose {
		fmt.Printf("✓ Read %d bytes\n\n", len(html))
	}

	// Detect parser from HTML content
	var parser Parser
	var siteName string

	if strings.Contains(html, "discogs.com") {
		parser = scraping.NewDiscogsParser()
		siteName = "Discogs"
	} else if strings.Contains(html, "prestomusic.com") || strings.Contains(html, "prestoclassical") {
		parser = scraping.NewPrestoParser()
		siteName = "Presto Classical"
	} else {
		fmt.Fprintf(os.Stderr, "Error: Cannot detect site from HTML file\n")
		fmt.Fprintf(os.Stderr, "File must be from a supported site:\n")
		fmt.Fprintf(os.Stderr, "  - Discogs\n")
		fmt.Fprintf(os.Stderr, "  - Presto Classical/Music\n")
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Detected site: %s\n", siteName)
	}

	// Parse HTML
	fmt.Printf("Parsing metadata...\n")
	result, err := parser.Parse(html)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML: %v\n", err)
		os.Exit(1)
	}

	// Process and output
	processResult(result)
}

// extractFromURL extracts metadata from a website
func extractFromURL() {
	// Determine which parser to use based on URL
	if *verbose {
		fmt.Printf("Analyzing URL: %s\n", *url)
	}

	var parser Parser
	var siteName string

	if isDiscogs(*url) {
		parser = scraping.NewDiscogsParser()
		siteName = "Discogs"
	} else if isPrestoClassical(*url) {
		parser = scraping.NewPrestoParser()
		siteName = "Presto Classical"
	} else {
		fmt.Fprintf(os.Stderr, "Error: No parser available for this URL\n")
		fmt.Fprintf(os.Stderr, "Supported sites:\n")
		fmt.Fprintf(os.Stderr, "  - discogs.com\n")
		fmt.Fprintf(os.Stderr, "  - prestoclassical.co.uk / prestomusic.com\n")
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
	processResult(result)
}

// processResult handles the common result processing and output with force mode support
func processResult(result *scraping.ExtractionResult) {
	data := result.Album

	// Display extraction summary
	fmt.Fprintf(os.Stderr, "✓ Extracted: %s", data.Title)
	if data.OriginalYear > 0 {
		fmt.Fprintf(os.Stderr, " (%d)", data.OriginalYear)
	}
	fmt.Fprintf(os.Stderr, "\n")

	if data.Edition != nil {
		if data.Edition.Label != "" {
			fmt.Fprintf(os.Stderr, "  Label: %s", data.Edition.Label)
			if data.Edition.CatalogNumber != "" {
				fmt.Fprintf(os.Stderr, " %s", data.Edition.CatalogNumber)
			}
			fmt.Println()
		} else if data.Edition.CatalogNumber != "" {
			fmt.Fprintf(os.Stderr, "  Catalog: %s\n", data.Edition.CatalogNumber)
		}
	}
	fmt.Fprintf(os.Stderr, "  Tracks: %d\n", len(data.Tracks))

	// Show extraction errors and warnings
	if len(result.Errors) > 0 {
		fmt.Println("\n⚠ Extraction Issues:")
		for _, e := range result.Errors {
			if e.Required {
				fmt.Fprintf(os.Stderr, "  ERROR: %s - %s\n", e.Field, e.Message)
			} else {
				fmt.Fprintf(os.Stderr, "  WARNING: %s - %s\n", e.Field, e.Message)
			}
		}
	}

	for _, w := range result.Warnings {
		fmt.Fprintf(os.Stderr, "  WARNING: %s\n", w)
	}

	// Fail if required errors and not forced
	if len(result.Errors) > 0 && !*force {
		fmt.Fprintf(os.Stderr, "\n❌ ERROR: Extraction failed due to required field errors\n")
		fmt.Fprintf(os.Stderr, "Use -force to create output anyway (not recommended for tagging)\n")
		os.Exit(1)
	}

	// Show force mode warning if used with errors
	if len(result.Errors) > 0 && *force {
		fmt.Fprintf(os.Stderr, "\n⚠️ WARNING: Forced output despite required field errors")
		fmt.Fprintf(os.Stderr, "This metadata may be incomplete and unsuitable for tagging")
	}

	// Convert to domain model and validate
	validationErrors := data.Validate()
	hasRequiredErrors := false
	for _, validationError := range validationErrors {
		switch validationError.Level {
		case domain.LevelError:
			if validationError.Required {
				hasRequiredErrors = true
			}
			fmt.Fprintf(os.Stderr, "\n❌ ERROR: Domain conversion failed: %v\n", validationError.Message)
		case domain.LevelWarning:
			fmt.Fprintf(os.Stderr, "\n⚠️ WARNING: %s\n", validationError.Message)
		default:
			fmt.Fprintf(os.Stderr, "\nℹ️ INFO: %s\n", validationError.Message)
		}
	}
	if hasRequiredErrors {
		if !*force {
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "⚠ WARNING: Continuing with partial data due to -force")
		}
	}

	if len(validationErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠ Validation warnings:")
		for _, verr := range validationErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", verr)
		}
	}

	// Serialize to JSON
	outFile := os.Stdout
	if *outputFile != "" {
		var err error
		outFile, err = os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
	}
	enc := json.NewEncoder(outFile)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	// Show verbose parsing notes
	if *verbose && len(result.Notes) > 0 {
		fmt.Fprintf(os.Stderr, "\nParsing Notes:")
		notesJSON, _ := json.MarshalIndent(result.Notes, "  ", "  ")
		fmt.Fprintf(os.Stderr, "  %s\n", string(notesJSON))
	}

	// Show next steps
	if !hasRequiredErrors {
		fmt.Fprintf(os.Stderr, "\nNext steps:")
		fmt.Fprintf(os.Stderr, "  1. Review Metadata: cat %s\n", *outputFile)
		fmt.Fprintf(os.Stderr, "  2. Validate album directory: validate /path/to/album\n")
		fmt.Fprintf(os.Stderr, "  3. Apply tags: tag -metadata %s -dir /path/to/album\n", *outputFile)
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

// isDiscogs checks if URL is from Discogs
func isDiscogs(url string) bool {
	return contains(url, "discogs.com/release/")
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
