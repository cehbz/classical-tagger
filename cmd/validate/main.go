package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/storage"
	"github.com/cehbz/classical-tagger/internal/validation"
)

// ValidationReport contains all validation results
type ValidationReport struct {
	MetadataFile  string
	ReferenceFile string
	Issues        []domain.ValidationIssue
	Torrent       *domain.Torrent
	LoadErrors    []error
}

// HasErrors returns true if there are any ERROR level issues
func (r *ValidationReport) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Level == domain.LevelError {
			return true
		}
	}
	return len(r.LoadErrors) > 0
}

// HasWarnings returns true if there are any WARNING level issues
func (r *ValidationReport) HasWarnings() bool {
	for _, issue := range r.Issues {
		if issue.Level == domain.LevelWarning {
			return true
		}
	}
	return false
}

// ValidateJSONFiles validates a JSON metadata file against validation rules.
// Optionally validates against a reference JSON file if provided.
func ValidateJSONFiles(metadataFile string, referenceFile string) (*ValidationReport, error) {
	report := &ValidationReport{
		MetadataFile:  metadataFile,
		ReferenceFile: referenceFile,
	}

	// Load JSON metadata file
	repo := storage.NewRepository()
	torrent, err := repo.LoadFromFile(metadataFile)
	if err != nil {
		report.LoadErrors = append(report.LoadErrors, fmt.Errorf("failed to load JSON metadata file: %w", err))
		// Torrent is nil when load fails - tests expect this behavior
		report.Torrent = nil
		return report, nil
	}
	report.Torrent = torrent

	// Load reference JSON file if provided
	var referenceTorrent *domain.Torrent
	if referenceFile != "" {
		refTorrent, err := repo.LoadFromFile(referenceFile)
		if err != nil {
			report.LoadErrors = append(report.LoadErrors, fmt.Errorf("failed to load reference JSON file: %w", err))
			// Continue validation without reference
		} else {
			referenceTorrent = refTorrent
		}
	}

	// Perform validation (only if torrent was loaded successfully)
	if torrent != nil {
		report.Issues = validation.Check(torrent, referenceTorrent)
	}

	return report, nil
}

// PrintReport formats and prints a validation report
func PrintReport(report *ValidationReport) {
	fmt.Printf("=== Validation Report ===\n\n")
	fmt.Printf("Metadata file: %s\n", report.MetadataFile)
	if report.ReferenceFile != "" {
		fmt.Printf("Reference file: %s\n", report.ReferenceFile)
	}
	fmt.Println()

	// Print load errors first
	if len(report.LoadErrors) > 0 {
		fmt.Println("❌ FILE LOAD ERRORS:")
		for _, err := range report.LoadErrors {
			fmt.Printf("  %v\n", err)
		}
		fmt.Println()
	}

	// Print validation issues
	if len(report.Issues) > 0 {
		fmt.Println("🏷️  VALIDATION ISSUES:")
		printIssues(report.Issues)
		fmt.Println()
	}

	// Summary
	fmt.Println("=== SUMMARY ===")
	if report.HasErrors() {
		fmt.Println("❌ FAILED: Album has critical errors")
	} else if report.HasWarnings() {
		fmt.Println("⚠️  WARNING: Album has warnings but is usable")
	} else {
		fmt.Println("✅ PASSED: Album is fully compliant")
	}

	fmt.Printf("  Issues: %d\n", len(report.Issues))
	fmt.Printf("  Load errors: %d\n", len(report.LoadErrors))
}

func printIssues(issues []domain.ValidationIssue) {
	for _, issue := range issues {
		symbol := "  "
		switch issue.Level {
		case domain.LevelError:
			symbol = "❌"
		case domain.LevelWarning:
			symbol = "⚠️ "
		case domain.LevelInfo:
			symbol = "ℹ️ "
		}
		fmt.Printf("%s %s\n", symbol, issue)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: validate <metadata.json> [reference.json]\n\n")
	fmt.Fprintf(os.Stderr, "Validates a JSON metadata file against validation rules.\n")
	fmt.Fprintf(os.Stderr, "If a reference JSON file is provided, validates against it as well.\n\n")
	fmt.Fprintf(os.Stderr, "Arguments:\n")
	fmt.Fprintf(os.Stderr, "  metadata.json   Required: Path to the JSON metadata file to validate\n")
	fmt.Fprintf(os.Stderr, "  reference.json  Optional: Path to a reference JSON file for comparison\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  # Validate a JSON metadata file:\n")
	fmt.Fprintf(os.Stderr, "  validate album.json\n\n")
	fmt.Fprintf(os.Stderr, "  # Validate against a reference:\n")
	fmt.Fprintf(os.Stderr, "  validate album.json reference.json\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: JSON metadata file is required\n\n")
		usage()
		os.Exit(1)
	}

	if flag.NArg() > 2 {
		fmt.Fprintf(os.Stderr, "Error: too many arguments\n\n")
		usage()
		os.Exit(1)
	}

	metadataFile := flag.Arg(0)
	referenceFile := ""
	if flag.NArg() == 2 {
		referenceFile = flag.Arg(1)
	}

	// Validate metadata file exists
	info, err := os.Stat(metadataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: metadata file '%s' not found: %v\n", metadataFile, err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: '%s' is a directory, expected a JSON file\n", metadataFile)
		os.Exit(1)
	}

	// Validate reference file exists if provided
	if referenceFile != "" {
		refInfo, err := os.Stat(referenceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: reference file '%s' not found: %v\n", referenceFile, err)
			os.Exit(1)
		}
		if refInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: '%s' is a directory, expected a JSON file\n", referenceFile)
			os.Exit(1)
		}
	}

	// Perform validation
	report, err := ValidateJSONFiles(metadataFile, referenceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}

	// Print report
	PrintReport(report)

	// Exit with error code if there are errors
	if report.HasErrors() {
		os.Exit(1)
	}
}
