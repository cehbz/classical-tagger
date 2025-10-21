package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/filesystem"
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
	dirValidator *filesystem.DirectoryValidator
}

// NewDirectoryScanner creates a new scanner
func NewDirectoryScanner() *DirectoryScanner {
	return &DirectoryScanner{
		dirValidator: filesystem.NewDirectoryValidator(),
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
		if issue.Level() == domain.LevelError {
			return true
		}
	}
	for _, issue := range r.MetadataIssues {
		if issue.Level() == domain.LevelError {
			return true
		}
	}
	return len(r.ReadErrors) > 0
}

// HasWarnings returns true if there are any WARNING level issues
func (r *ValidationReport) HasWarnings() bool {
	for _, issue := range r.StructureIssues {
		if issue.Level() == domain.LevelWarning {
			return true
		}
	}
	for _, issue := range r.MetadataIssues {
		if issue.Level() == domain.LevelWarning {
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
		
		// Read track from file
		track, err := reader.ReadTrackFromFile(file, discNum, 0) // track number will be read from tags
		if err != nil {
			report.ReadErrors = append(report.ReadErrors, fmt.Errorf("%s: %w", file, err))
			continue
		}
		
		// Build album on first track
		if album == nil {
			// Read album-level metadata from first file
			metadata, err := reader.ReadFile(file)
			if err != nil {
				report.ReadErrors = append(report.ReadErrors, fmt.Errorf("read album metadata: %w", err))
				continue
			}
			
			album, err = domain.NewAlbum(metadata.Album, 0) // year will be validated separately
			if err != nil {
				report.ReadErrors = append(report.ReadErrors, fmt.Errorf("create album: %w", err))
				continue
			}
		}
		
		// Add track to album
		err = album.AddTrack(track)
		if err != nil {
			report.ReadErrors = append(report.ReadErrors, fmt.Errorf("add track: %w", err))
			continue
		}
	}
	
	report.Album = album
	
	// Validate metadata if we successfully built an album
	if album != nil {
		validator := validation.NewAlbumValidator()
		report.MetadataIssues = validator.ValidateMetadata(album)
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
		switch issue.Level() {
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
