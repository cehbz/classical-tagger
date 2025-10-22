package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseLeadingTrackNumber(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
		ok   bool
	}{
		{"two-digit with dash", "01 - Title.flac", 1, true},
		{"multi-digit no dash", "123 Title.flac", 123, true},
		{"single digit with dot", "7. Foo.flac", 7, true},
		{"no number", "Foo.flac", 0, false},
		{"starts with word then number", "Track 01 - Title.flac", 0, false}, // TODO: this maybe should pass if all tracks have same word prefix
		{"starts with space then number", " 5 - Title.flac", 5, true},
		{"number then dot no ext", "10. Title", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLeadingTrackNumber(tt.in)
			if tt.ok && err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if tt.ok && got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseYearFromFolderName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"leading YYYY dash", "2019 - Album Title [FLAC]", 2019},
		{"paren year", "Artist - Album (1974) [FLAC]", 1974},
		{"multiple paren take latest", "Artist - X (1993 Remaster) (2013) [FLAC]", 2013},
		{"latest year earlier in name", "2018 Remaster (1993) - Album (2013)", 2018},
		{"any 4-digit fallback", "Band - Title - 2009 - FLAC", 2009},
		{"no year", "Band - Title - FLAC", 0},
		{"year in braces", "Album {5054197044984} (2019) [FLAC]", 2019},
		{"weird unicode", "Noël. Christmas. Weinachten - RIAS Kammerchor - Rademann [96_24]", 0},
		{"future year rejected", "Artist - Title (2999)", 0},
		{"too early year rejected", "Artist - Title (1899)", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseYearFromFolderName(tt.in)
			if got != tt.want {
				t.Fatalf("parseYearFromFolderName(%q)=%d, want %d", tt.in, got, tt.want)
			}
		})
	}

	// Also verify that current year is accepted
	current := time.Now().Year()
	name := strings.Replace("ART - NAME (YYYY)", "YYYY", fmt.Sprintf("%d", current), 1)
	got := parseYearFromFolderName(name)
	if got != current {
		t.Fatalf("expected current year %d, got %d", current, got)
	}
}

// TestValidateCommand tests the main validation flow
func TestValidateCommand(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create a simple single-disc structure
	albumDir := filepath.Join(tmpDir, "Bach - Goldberg Variations (1981) - FLAC")
	err := os.MkdirAll(albumDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test FLAC file (empty is fine for structure validation)
	testFile := filepath.Join(albumDir, "01 Aria.flac")
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()

	// Test validation
	scanner := NewDirectoryScanner()
	structure, err := scanner.Scan(albumDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if structure.BasePath != albumDir {
		t.Errorf("BasePath = %q, want %q", structure.BasePath, albumDir)
	}

	if len(structure.Files) != 1 {
		t.Errorf("Files count = %d, want 1", len(structure.Files))
	}
}

func TestValidateMultiDisc(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-disc structure
	albumDir := filepath.Join(tmpDir, "Album - Title (2020) - FLAC")
	cd1Dir := filepath.Join(albumDir, "CD1")
	cd2Dir := filepath.Join(albumDir, "CD2")

	err := os.MkdirAll(cd1Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create CD1: %v", err)
	}
	err = os.MkdirAll(cd2Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create CD2: %v", err)
	}

	// Create test files
	os.Create(filepath.Join(cd1Dir, "01 Track.flac"))
	os.Create(filepath.Join(cd2Dir, "01 Track.flac"))

	scanner := NewDirectoryScanner()
	structure, err := scanner.Scan(albumDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !structure.IsMultiDisc {
		t.Error("Expected multi-disc structure")
	}

	if len(structure.Files) != 2 {
		t.Errorf("Files count = %d, want 2", len(structure.Files))
	}
}

func TestValidatePathLength(t *testing.T) {
	scanner := NewDirectoryScanner()

	// Create a very long path
	longPath := "/" + string(make([]byte, 190))
	for i := range longPath {
		if longPath[i] == 0 {
			longPath = longPath[:i] + "a" + longPath[i+1:]
		}
	}

	structure := &DirectoryStructure{
		BasePath: longPath,
		Files:    []string{longPath + "/file.flac"},
	}

	issues := scanner.ValidateStructure(structure)

	// Should have at least one error about path length
	hasPathError := false
	for _, issue := range issues {
		if issue.Track() == -1 { // directory-level issue
			hasPathError = true
			break
		}
	}

	if !hasPathError {
		t.Error("Expected path length validation error")
	}
}

func TestValidateDirectory_NoTrackNumberInFilename(t *testing.T) {
	tmpDir := t.TempDir()
	albumDir := filepath.Join(tmpDir, "Album - Title (2020) - FLAC")
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create file without leading track number
	f := filepath.Join(albumDir, "Track Title.flac")
	if err := os.WriteFile(f, []byte{}, 0644); err != nil {
		t.Fatalf("create: %v", err)
	}

	report, err := ValidateDirectory(albumDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}
	if len(report.ReadErrors) == 0 {
		t.Fatalf("expected read error for missing filename track number")
	}
	found := false
	for _, e := range report.ReadErrors {
		if strings.Contains(e.Error(), "2.3.13: filename must start with track number") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 2.3.13 error, got: %v", report.ReadErrors[0])
	}
}

func TestValidateDirectory_TrackNumberButNoTag(t *testing.T) {
	tmpDir := t.TempDir()
	albumDir := filepath.Join(tmpDir, "Album - Title (2020) - FLAC")
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// File has leading track number but empty content (no tags)
	f := filepath.Join(albumDir, "01 Track Title.flac")
	if err := os.WriteFile(f, []byte{}, 0644); err != nil {
		t.Fatalf("create: %v", err)
	}

	report, err := ValidateDirectory(albumDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}
	if len(report.ReadErrors) == 0 {
		t.Fatalf("expected read error due to missing tags")
	}
}
