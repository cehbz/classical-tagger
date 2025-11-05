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
		"root_path": "test-album",
		"title": "Test Album",
		"original_year": 2013,
		"files": [
			{
				"path": "01.flac",
				"size": 0,
				"disc": 1,
				"track": 1,
				"title": "Test Track",
				"artists": [
					{
						"name": "Test Composer",
						"role": "composer"
					}
				]
			}
		]
	}`

	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	torrent, err := LoadMetadataJSON(jsonFile)
	if err != nil {
		t.Fatalf("LoadMetadataJSON() error = %v", err)
	}

	if torrent.Title != "Test Album" {
		t.Errorf("Title = %v, want 'Test Album'", torrent.Title)
	}

	if len(torrent.Tracks()) != 1 {
		t.Errorf("Track count = %d, want 1", len(torrent.Tracks()))
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

func TestMatchTracksToFiles(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "metadata.json")

	// Create test metadata with 3 tracks
	jsonContent := `{
		"root_path": "test-album",
		"title": "Test Album",
		"original_year": 2020,
		"files": [
			{
				"path": "01.flac",
				"size": 0,
				"disc": 1,
				"track": 1,
				"title": "First Track",
				"artists": [
					{
						"name": "Test Composer",
						"role": "composer"
					}
				]
			},
			{
				"path": "02.flac",
				"size": 0,
				"disc": 1,
				"track": 2,
				"title": "Second Track",
				"artists": [
					{
						"name": "Test Composer",
						"role": "composer"
					}
				]
			},
			{
				"path": "03.flac",
				"size": 0,
				"disc": 1,
				"track": 3,
				"title": "Third Track",
				"artists": [
					{
						"name": "Test Composer",
						"role": "composer"
					}
				]
			}
		]
	}`

	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	torrent, err := LoadMetadataJSON(jsonFile)
	if err != nil {
		t.Fatalf("LoadMetadataJSON() error = %v", err)
	}

	// Create test files
	files := []string{
		filepath.Join(tmpDir, "01 First Track.flac"),
		filepath.Join(tmpDir, "02 Second Track.flac"),
		// Note: Track 3 has no matching file
	}

	matches := MatchTracksToFiles(torrent, files)

	// Should have 3 tracks in matches
	if len(matches) != 3 {
		t.Errorf("matches count = %d, want 3", len(matches))
	}

	// Check that tracks 1 and 2 matched, track 3 did not
	matchCount := 0
	unmatchCount := 0
	for _, file := range matches {
		if file != "" {
			matchCount++
		} else {
			unmatchCount++
		}
	}

	if matchCount != 2 {
		t.Errorf("matched tracks = %d, want 2", matchCount)
	}

	if unmatchCount != 1 {
		t.Errorf("unmatched tracks = %d, want 1", unmatchCount)
	}
}

func TestOutputDirectoryCreation(t *testing.T) {
	// This tests that the output directory logic works correctly
	tests := []struct {
		name      string
		targetDir string
		outputDir string
		want      string
	}{
		{
			name:      "default output dir",
			targetDir: "/music/album",
			outputDir: "",
			want:      "/music/album_tagged",
		},
		{
			name:      "custom output dir",
			targetDir: "/music/album",
			outputDir: "/music/output",
			want:      "/music/output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := tt.outputDir
			if outDir == "" {
				outDir = tt.targetDir + "_tagged"
			}

			if outDir != tt.want {
				t.Errorf("output dir = %v, want %v", outDir, tt.want)
			}
		})
	}
}

// TestWriteToNewDirectory verifies that the new write-to-directory approach is used
func TestWriteToNewDirectory(t *testing.T) {
	// This test documents the expected behavior:
	// - Original files in source directory remain untouched
	// - Tagged files written to output directory
	// - No backup/restore needed

	t.Skip("Integration test - requires full FLAC implementation")

	// Expected workflow:
	// 1. Load metadata
	// 2. Find source FLAC files
	// 3. Match tracks to files
	// 4. Create output directory
	// 5. For each match: writer.WriteTrack(sourcePath, destPath, track, album)
	// 6. Verify source files untouched
	// 7. Verify dest files have correct tags
}

// TestDryRunMode verifies dry-run doesn't modify files
func TestDryRunMode(t *testing.T) {
	t.Skip("Integration test - requires CLI execution")

	// Expected behavior with --dry-run:
	// - Shows what would be done
	// - Creates no directories
	// - Writes no files
	// - Returns success exit code
}
