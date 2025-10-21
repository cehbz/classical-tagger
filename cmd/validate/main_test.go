package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
