package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/tagging"
	"github.com/cehbz/classical-tagger/internal/validation"
)

// DirectoryStructure represents a scanned album directory
type DirectoryStructure struct {
	BasePath    string
	FolderName  string
	Files       []string
	IsMultiDisc bool
}

// DirectoryScanner scans album directories
type DirectoryScanner struct {
	dirValidator *validation.DirectoryValidator
}

// NewDirectoryScanner creates a new scanner
func NewDirectoryScanner() *DirectoryScanner {
	return &DirectoryScanner{
		dirValidator: validation.NewDirectoryValidator(),
	}
}

// Scan recursively scans a directory for FLAC files
func (s *DirectoryScanner) Scan(basePath string) (*DirectoryStructure, error) {
	structure := &DirectoryStructure{
		BasePath:   basePath,
		FolderName: filepath.Base(basePath),
		Files:      []string{},
	}

	// Walk the directory tree
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory itself
		if path == basePath {
			return nil
		}

		// Check for disc directories
		if info.IsDir() {
			relPath := strings.TrimPrefix(path, basePath)
			relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

			// Check if this is a disc directory (CD1, CD2, etc.)
			if isDiscDirectory(filepath.Base(path)) {
				structure.IsMultiDisc = true
			}

			return nil
		}

		// Only collect FLAC files
		if strings.HasSuffix(strings.ToLower(path), ".flac") {
			structure.Files = append(structure.Files, path)
		}

		return nil
	})

	return structure, err
}

// ValidateStructure validates directory structure and paths
func (s *DirectoryScanner) ValidateStructure(structure *DirectoryStructure) []domain.ValidationIssue {
	var issues []domain.ValidationIssue

	// Validate folder name
	// Note: We can't validate against an album object here since we haven't read tags yet
	// Just check basic path rules
	folderIssues := s.dirValidator.ValidatePath(structure.BasePath)
	issues = append(issues, folderIssues...)

	// Validate file paths
	for _, file := range structure.Files {
		pathIssues := s.dirValidator.ValidatePath(file)
		issues = append(issues, pathIssues...)
	}

	// Validate overall structure (multi-disc detection)
	structureIssues := s.dirValidator.ValidateStructure(structure.BasePath, structure.Files)
	issues = append(issues, structureIssues...)

	return issues
}

// isDiscDirectory checks if a directory name looks like a disc folder
func isDiscDirectory(name string) bool {
	lower := strings.ToLower(name)
	patterns := []string{"cd", "disc", "disk"}

	for _, pattern := range patterns {
		if strings.HasPrefix(lower, pattern) {
			rest := strings.TrimPrefix(lower, pattern)
			rest = strings.TrimSpace(rest)
			if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
				return true
			}
		}
	}

	return false
}

// parseLeadingTrackNumber extracts the leading integer from a filename (before extension).
// Complies with rules 2.3.13 and 2.3.14 requiring track numbers at the start.
func parseLeadingTrackNumber(name string) (int, error) {
	numStr := strings.TrimSpace(name)
	i := 0
	for ; i < len(numStr); i++ {
		if numStr[i] < '0' || numStr[i] > '9' {
			break
		}
	}
	if i == 0 {
		return 0, fmt.Errorf("2.3.13: filename must start with track number, got %q", name)
	}
	numStr = numStr[:i]
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid leading track number %q in %q", numStr, name)
	}
	return n, nil
}

// parseYearFromFolderName tries to extract a plausible release year from the album folder name.
// Heuristics, based on examples:
// - Prefer a leading "YYYY -" pattern
// - Else prefer the first parenthesized year ("(YYYY)")
// - Else take the first 4-digit year found
// Only accept years in [1900, currentYear+1]. Returns 0 if none found.
func parseYearFromFolderName(folder string) int {
	yearRangeMin := 1900
	yearRangeMax := time.Now().Year()

	// Collect all 4-digit sequences and return the latest valid one
	anyRe := regexp.MustCompile(`\b(\d{4})\b`)
	matches := anyRe.FindAllStringSubmatch(folder, -1)
	best := 0
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		if y, err := strconv.Atoi(m[1]); err == nil && y >= yearRangeMin && y <= yearRangeMax {
			if y > best {
				best = y
			}
		}
	}
	return best
}

// ValidationReport contains all validation results
type ValidationReport struct {
	Path            string
	StructureIssues []domain.ValidationIssue
	MetadataIssues  []domain.ValidationIssue
	Album           *domain.Album
	ReadErrors      []error
}

// HasErrors returns true if there are any ERROR level issues
func (r *ValidationReport) HasErrors() bool {
	for _, issue := range r.StructureIssues {
		if issue.Level == domain.LevelError {
			return true
		}
	}
	for _, issue := range r.MetadataIssues {
		if issue.Level == domain.LevelError {
			return true
		}
	}
	return len(r.ReadErrors) > 0
}

// HasWarnings returns true if there are any WARNING level issues
func (r *ValidationReport) HasWarnings() bool {
	for _, issue := range r.StructureIssues {
		if issue.Level == domain.LevelWarning {
			return true
		}
	}
	for _, issue := range r.MetadataIssues {
		if issue.Level == domain.LevelWarning {
			return true
		}
	}
	return false
}

// ValidateDirectory performs complete validation of a directory
func ValidateDirectory(path string) (*ValidationReport, error) {
	report := &ValidationReport{
		Path: path,
	}

	// Scan directory structure
	scanner := NewDirectoryScanner()
	structure, err := scanner.Scan(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// Validate structure
	report.StructureIssues = scanner.ValidateStructure(structure)

	// Read FLAC tags and build album
	reader := tagging.NewFLACReader()
	var album *domain.Album

	for _, file := range structure.Files {
		// Determine disc and track number from file path
		relPath := strings.TrimPrefix(file, structure.BasePath)
		relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

		discNum := 1
		if structure.IsMultiDisc {
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) > 1 && isDiscDirectory(parts[0]) {
				// Extract disc number from directory name (e.g., "CD1" -> 1)
				discName := strings.ToLower(parts[0])
				for _, prefix := range []string{"cd", "disc", "disk"} {
					if strings.HasPrefix(discName, prefix) {
						discName = strings.TrimPrefix(discName, prefix)
						break
					}
				}
				fmt.Sscanf(discName, "%d", &discNum)
			}
		}

		// Derive expected track number from filename per rules (2.3.13/2.3.14)
		filename := filepath.Base(file)
		expectedTrack, parseErr := parseLeadingTrackNumber(filename)
		if parseErr != nil {
			report.ReadErrors = append(report.ReadErrors, fmt.Errorf("%s: %v", file, parseErr))
			continue
		}

		// Read track from file and validate against expected disc/track
		track, err := reader.ReadTrackFromFile(file, discNum, expectedTrack)
		if err != nil {
			report.ReadErrors = append(report.ReadErrors, fmt.Errorf("%s: %w", file, err))
			continue
		}

		// Build album on first track
		if album == nil {
			// Read album-level metadata from first file
			metadata, err := reader.ReadFile(file)
			if err != nil {
				report.ReadErrors = append(report.ReadErrors, fmt.Errorf("read album Metadata: %w", err))
				continue
			}

			// Parse year from tags if present and sane; else fallback to folder name; allow 0 if unknown
			originalYear := 0
			if y, err := strconv.Atoi(strings.TrimSpace(metadata.Year)); err == nil && y >= 0 {
				current := time.Now().Year()
				if y >= 1900 && y <= current {
					originalYear = y
				}
			}
			if originalYear == 0 {
				originalYear = parseYearFromFolderName(structure.FolderName)
			}

			album = &domain.Album{Title: metadata.Album, OriginalYear: originalYear}
		}

		// Add track to album
		album.Tracks = append(album.Tracks, track)
		if err != nil {
			report.ReadErrors = append(report.ReadErrors, fmt.Errorf("add track: %w", err))
			continue
		}
	}

	report.Album = album

	// Validate metadata if we successfully built an album
	if album != nil {
		report.MetadataIssues = validation.Check(album, nil)

		// Validate folder name against album metadata
		// (requires album metadata to check composer name, etc.)
		folderIssues := scanner.dirValidator.ValidateFolderName(structure.FolderName, album)
		report.StructureIssues = append(report.StructureIssues, folderIssues...)
	}

	return report, nil
}

// PrintReport formats and prints a validation report
func PrintReport(report *ValidationReport) {
	fmt.Printf("=== Validation Report: %s ===\n\n", report.Path)

	// Print read errors first
	if len(report.ReadErrors) > 0 {
		fmt.Println("❌ FILE READ ERRORS:")
		for _, err := range report.ReadErrors {
			fmt.Printf("  %v\n", err)
		}
		fmt.Println()
	}

	// Print structure issues
	if len(report.StructureIssues) > 0 {
		fmt.Println("📁 DIRECTORY STRUCTURE ISSUES:")
		printIssues(report.StructureIssues)
		fmt.Println()
	}

	// Print metadata issues
	if len(report.MetadataIssues) > 0 {
		fmt.Println("🏷️  METADATA ISSUES:")
		printIssues(report.MetadataIssues)
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

	fmt.Printf("  Structure issues: %d\n", len(report.StructureIssues))
	fmt.Printf("  Metadata issues: %d\n", len(report.MetadataIssues))
	fmt.Printf("  Read errors: %d\n", len(report.ReadErrors))
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

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: validate <directory>\n")
		os.Exit(1)
	}

	path := flag.Arg(0)

	// Validate path exists
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", path)
		os.Exit(1)
	}

	// Perform validation
	report, err := ValidateDirectory(path)
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
