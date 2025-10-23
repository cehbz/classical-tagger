package scraping

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalExtractor_ExtractTrackNumberFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     int
	}{
		{"01 Prelude.flac", 1},
		{"01-Prelude.flac", 1},
		{"01.Prelude.flac", 1},
		{"01_Prelude.flac", 1},
		{"1 Prelude.flac", 1},
		{"123 Track.flac", 123},
		{"Prelude.flac", 0}, // No number
		{"abc01.flac", 0},   // Number not at start
	}

	extractor := NewLocalExtractor()
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := extractor.extractTrackNumberFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("extractTrackNumberFromFilename(%q) = %d, want %d", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLocalExtractor_ExtractDiscFromPath(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{"/music/album/CD1/01.flac", 1},
		{"/music/album/CD2/01.flac", 2},
		{"/music/album/Disc 1/01.flac", 1},
		{"/music/album/Disc 2/01.flac", 2},
		{"/music/album/disc1/01.flac", 1},
		{"/music/album/01.flac", 1}, // No disc indicator, default to 1
	}

	extractor := NewLocalExtractor()
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractor.extractDiscFromPath(tt.path)
			if got != tt.want {
				t.Errorf("extractDiscFromPath(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}

func TestLocalExtractor_ExtractTitleFromFilename(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"01 Prelude.flac", "Prelude"},
		{"01-Prelude in G major.flac", "Prelude in G major"},
		{"01.Track Title.flac", "Track Title"},
		{"01_Some Title.flac", "Some Title"},
		{"123 Long Title Here.flac", "Long Title Here"},
		{"Title without number.flac", "Title without number"},
	}

	extractor := NewLocalExtractor()
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractor.extractTitleFromFilename(tt.path)
			if got != tt.want {
				t.Errorf("extractTitleFromFilename(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestLocalExtractor_ParseArtistField(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  int // Number of artists
	}{
		{
			name:  "semicolon separated",
			field: "Martha Argerich; Berlin Philharmonic Orchestra; Claudio Abbado",
			want:  3,
		},
		{
			name:  "comma separated",
			field: "Martha Argerich, Berlin Philharmonic Orchestra, Claudio Abbado",
			want:  3,
		},
		{
			name:  "single artist",
			field: "Glenn Gould",
			want:  1,
		},
		{
			name:  "empty",
			field: "",
			want:  0,
		},
	}

	extractor := NewLocalExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.parseArtistField(tt.field)
			if len(got) != tt.want {
				t.Errorf("parseArtistField(%q) returned %d artists, want %d", tt.field, len(got), tt.want)
			}
		})
	}
}

func TestLocalExtractor_InferRoleFromName(t *testing.T) {
	tests := []struct {
		name     string
		wantRole string
	}{
		{"Herbert von Karajan, conductor", "conductor"},
		{"Berlin Philharmonic Orchestra", "ensemble"},
		{"Emerson String Quartet", "ensemble"},
		{"Martha Argerich", "soloist"},
		{"John Doe", "soloist"}, // Default
	}

	extractor := NewLocalExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.inferRoleFromName(tt.name)
			if got != tt.wantRole {
				t.Errorf("inferRoleFromName(%q) = %q, want %q", tt.name, got, tt.wantRole)
			}
		})
	}
}

func TestLocalExtractor_ParseDirectoryName(t *testing.T) {
	tests := []struct {
		dirName   string
		wantTitle string
		wantYear  int
	}{
		{
			dirName:   "Bach - Goldberg Variations (1981) - FLAC",
			wantTitle: "Bach - Goldberg Variations",
			wantYear:  1981,
		},
		{
			dirName:   "Beethoven - Symphony No. 9 (1989)",
			wantTitle: "Beethoven - Symphony No. 9",
			wantYear:  1989,
		},
		{
			dirName:   "Mozart - Piano Concertos (2005) - 24-96",
			wantTitle: "Mozart - Piano Concertos",
			wantYear:  2005,
		},
		{
			dirName:   "Some Album",
			wantTitle: "Some Album",
			wantYear:  0,
		},
	}

	extractor := NewLocalExtractor()
	for _, tt := range tests {
		t.Run(tt.dirName, func(t *testing.T) {
			gotTitle, gotYear := extractor.parseDirectoryName(tt.dirName)
			if gotTitle != tt.wantTitle {
				t.Errorf("parseDirectoryName(%q) title = %q, want %q", tt.dirName, gotTitle, tt.wantTitle)
			}
			if gotYear != tt.wantYear {
				t.Errorf("parseDirectoryName(%q) year = %d, want %d", tt.dirName, gotYear, tt.wantYear)
			}
		})
	}
}

func TestLocalExtractor_FindFLACFiles(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"01.flac",
		"02.flac",
		"subdir/03.flac",
		"other.txt", // Should be ignored
		"test.FLAC", // Case insensitive
	}

	for _, f := range testFiles {
		fullPath := filepath.Join(tmpDir, f)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	extractor := NewLocalExtractor()
	files, err := extractor.findFLACFiles(tmpDir)

	if err != nil {
		t.Fatalf("findFLACFiles() error = %v", err)
	}

	// Should find 4 FLAC files (case insensitive)
	if len(files) != 4 {
		t.Errorf("findFLACFiles() found %d files, want 4", len(files))
	}

	// Verify all are FLAC files
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if ext != ".flac" {
			t.Errorf("findFLACFiles() returned non-FLAC file: %s", f)
		}
	}
}

func TestLocalExtractor_ExtractFromDirectory_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	extractor := NewLocalExtractor()
	_, err := extractor.ExtractFromDirectory(tmpDir)

	if err == nil {
		t.Error("ExtractFromDirectory() expected error for empty directory, got nil")
	}
}

func TestLocalExtractor_ExtractFromDirectory_InvalidPath(t *testing.T) {
	extractor := NewLocalExtractor()
	_, err := extractor.ExtractFromDirectory("/nonexistent/path")

	if err == nil {
		t.Error("ExtractFromDirectory() expected error for invalid path, got nil")
	}
}

// TestLocalExtractor_ExtractFromDirectory_RealDirectory is an integration test
// that requires real FLAC files to be present. It's skipped by default.
func TestLocalExtractor_ExtractFromDirectory_RealDirectory(t *testing.T) {
	// This test requires a real directory with FLAC files
	testDir := os.Getenv("TEST_FLAC_DIR")
	if testDir == "" {
		t.Skip("Set TEST_FLAC_DIR environment variable to test with real files")
	}

	extractor := NewLocalExtractor()
	result, err := extractor.ExtractFromDirectory(testDir)

	if err != nil {
		t.Fatalf("ExtractFromDirectory() error = %v", err)
	}

	data := result.Data()

	// Basic validations
	if data.Title == "" {
		t.Error("No album title extracted")
	}
	t.Logf("Album: %s", data.Title)

	if data.OriginalYear == 0 {
		t.Log("Warning: No year extracted")
	} else {
		t.Logf("Year: %d", data.OriginalYear)
	}

	if len(data.Tracks) == 0 {
		t.Error("No tracks extracted")
	}
	t.Logf("Tracks: %d", len(data.Tracks))

	// Validate tracks
	for i, track := range data.Tracks {
		if track.Track == 0 {
			t.Errorf("Track %d has no track number", i)
		}
		if track.Title == "" {
			t.Errorf("Track %d has no title", i)
		}
		t.Logf("  %d. %s", track.Track, track.Title)
	}

	// Check for errors
	if result.HasRequiredErrors() {
		t.Error("Extraction has required errors:")
		for _, e := range result.Errors() {
			if e.Required() {
				t.Errorf("  - %s: %s", e.Field(), e.Message())
			}
		}
	}

	// Try domain conversion
	_, err = data.ToAlbum()
	if err != nil {
		t.Logf("Domain conversion failed: %v", err)
	}
}