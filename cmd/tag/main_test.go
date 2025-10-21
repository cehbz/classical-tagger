package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMetadataJSON(t *testing.T) {
	// Create temp JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "metadata.json")
	
	jsonContent := `{
		"title": "Test Album",
		"original_year": 2013,
		"tracks": [
			{
				"disc": 1,
				"track": 1,
				"title": "Test Track",
				"composer": {
					"name": "Test Composer",
					"role": "composer"
				},
				"artists": [],
				"name": "01 Test Track.flac"
			}
		]
	}`
	
	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	album, err := LoadMetadataJSON(jsonFile)
	if err != nil {
		t.Fatalf("LoadMetadataJSON() error = %v", err)
	}
	
	if album.Title() != "Test Album" {
		t.Errorf("Title = %v, want 'Test Album'", album.Title())
	}
	
	if len(album.Tracks()) != 1 {
		t.Errorf("Track count = %d, want 1", len(album.Tracks()))
	}
}

func TestFindFLACFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test structure
	os.Create(filepath.Join(tmpDir, "01 Track.flac"))
	os.Create(filepath.Join(tmpDir, "02 Track.flac"))
	os.Create(filepath.Join(tmpDir, "cover.jpg")) // should be ignored
	
	files, err := FindFLACFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindFLACFiles() error = %v", err)
	}
	
	if len(files) != 2 {
		t.Errorf("Found %d files, want 2", len(files))
	}
}

func TestMatchTrackToFile(t *testing.T) {
	tests := []struct {
		name        string
		trackName   string
		files       []string
		wantMatch   bool
		wantFile    string
	}{
		{
			name:      "exact match",
			trackName: "01 Aria.flac",
			files:     []string{"01 Aria.flac", "02 Variation 1.flac"},
			wantMatch: true,
			wantFile:  "01 Aria.flac",
		},
		{
			name:      "no match",
			trackName: "03 Track.flac",
			files:     []string{"01 Track.flac", "02 Track.flac"},
			wantMatch: false,
		},
		{
			name:      "case insensitive match",
			trackName: "01 aria.flac",
			files:     []string{"01 Aria.flac"},
			wantMatch: true,
			wantFile:  "01 Aria.flac",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, file := MatchTrackToFile(tt.trackName, tt.files)
			
			if match != tt.wantMatch {
				t.Errorf("MatchTrackToFile() match = %v, want %v", match, tt.wantMatch)
			}
			
			if tt.wantMatch && file != tt.wantFile {
				t.Errorf("MatchTrackToFile() file = %v, want %v", file, tt.wantFile)
			}
		})
	}
}

func TestValidateBeforeApply(t *testing.T) {
	// This would test the validation that runs before applying tags
	// to ensure we don't corrupt files
	t.Skip("Integration test - requires full setup")
}
