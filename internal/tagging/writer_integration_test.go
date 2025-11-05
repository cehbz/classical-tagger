package tagging

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
)

// generateTestFLAC creates a minimal valid FLAC file for testing.
// Returns path to created file.
func generateTestFLAC(t *testing.T, dir string, filename string) string {
	t.Helper()

	// Create a 1-second sine wave at 440Hz (A4 note)
	// Using ffmpeg to generate minimal FLAC
	outPath := filepath.Join(dir, filename)

	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "sine=frequency=440:duration=1",
		"-ac", "2", // stereo
		"-ar", "44100", // 44.1kHz
		"-sample_fmt", "s16", // 16-bit
		"-y", // overwrite
		outPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to generate test FLAC: %v\nOutput: %s", err, output)
	}

	return outPath
}

// readVorbisComments reads all vorbis comments from a FLAC file.
func readVorbisComments(t *testing.T, path string) map[string]string {
	t.Helper()

	flacFile, err := flac.ParseFile(path)
	if err != nil {
		t.Fatalf("Failed to parse FLAC: %v", err)
	}

	tags := make(map[string]string)

	for _, metaBlock := range flacFile.Meta {
		if metaBlock.Type == flac.VorbisComment {
			cmtBlock, err := flacvorbis.ParseFromMetaDataBlock(*metaBlock)
			if err != nil {
				t.Fatalf("Failed to parse vorbis comment: %v", err)
			}

			for _, comment := range cmtBlock.Comments {
				// Comments are in "KEY=VALUE" format
				parts := splitComment(comment)
				if len(parts) == 2 {
					tags[parts[0]] = parts[1]
				}
			}
		}
	}

	return tags
}

// splitComment splits a vorbis comment "KEY=VALUE" string.
func splitComment(comment string) []string {
	for i := 0; i < len(comment); i++ {
		if comment[i] == '=' {
			return []string{comment[:i], comment[i+1:]}
		}
	}
	return []string{comment}
}

// TestFLACWriter_WriteTrack_Integration tests the complete write workflow.
func TestFLACWriter_WriteTrack_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Generate test FLAC
	sourcePath := generateTestFLAC(t, tmpDir, "source.flac")
	destPath := filepath.Join(tmpDir, "dest.flac")
	track := &domain.Track{
		File: domain.File{
			Path: "01.flac",
			Size: 0,
		},
		Disc:  1,
		Track: 1,
		Title: "Goldberg Variations, BWV 988: Aria",
		Artists: []domain.Artist{
			{Name: "Johann Sebastian Bach", Role: domain.RoleComposer},
			{Name: "Glenn Gould", Role: domain.RoleSoloist},
		},
	}

	// Create test metadata
	torrent := &domain.Torrent{
		RootPath:     "goldberg",
		Title:        "Goldberg Variations",
		OriginalYear: 1955,
		Edition: &domain.Edition{
			Label: "Sony Classical", Year: 1992, CatalogNumber: "SK 52594",
		},
		Files: []domain.FileLike{track},
	}

	// Write tags
	writer := NewFLACWriter()
	err := writer.WriteTrack(sourcePath, destPath, track, torrent)
	if err != nil {
		t.Fatalf("WriteTrack() error = %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatal("Destination file was not created")
	}

	// Verify tags written correctly
	tags := readVorbisComments(t, destPath)

	expectedTags := map[string]string{
		"TITLE":         "Goldberg Variations, BWV 988: Aria",
		"ALBUM":         "Goldberg Variations",
		"COMPOSER":      "Johann Sebastian Bach",
		"ARTIST":        "Glenn Gould",
		"PERFORMER":     "Glenn Gould",
		"TRACKNUMBER":   "1",
		"DISCNUMBER":    "1",
		"ORIGINALDATE":  "1955",
		"DATE":          "1992",
		"LABEL":         "Sony Classical",
		"CATALOGNUMBER": "SK 52594",
	}

	for key, want := range expectedTags {
		got, exists := tags[key]
		if !exists {
			t.Errorf("Missing tag %q", key)
			continue
		}
		if got != want {
			t.Errorf("Tag %q = %q, want %q", key, got, want)
		}
	}

	// Note: Audio preservation should be verified with metaflac --show-md5sum
	// in manual testing, as go-flac doesn't provide direct frame access
}

// TestFLACWriter_SpecialCharacters tests handling of Unicode and special chars.
func TestFLACWriter_SpecialCharacters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	sourcePath := generateTestFLAC(t, tmpDir, "source.flac")
	destPath := filepath.Join(tmpDir, "dest.flac")

	// Create metadata with special characters
	track := &domain.Track{
		File: domain.File{
			Path: "01.flac",
			Size: 0,
		},
		Disc:  1,
		Track: 1,
		Title: "Sonate für Violine und Klavier: Frisch (Œuvre)",
		Artists: []domain.Artist{
			{Name: "Béla Bartók", Role: domain.RoleComposer},
			{Name: "Mstislav Rostropóvič", Role: domain.RoleSoloist},
		},
	}
	torrent := &domain.Torrent{
		RootPath:     "bartok",
		Title:        "Bartók: Complete Works",
		OriginalYear: 2020,
		Edition: &domain.Edition{
			Label: "Sony Classical", Year: 1992, CatalogNumber: "SK 52594",
		},
		Files: []domain.FileLike{track},
	}

	writer := NewFLACWriter()
	err := writer.WriteTrack(sourcePath, destPath, track, torrent)
	if err != nil {
		t.Fatalf("WriteTrack() error = %v", err)
	}

	// Read back and verify
	tags := readVorbisComments(t, destPath)

	if tags["COMPOSER"] != "Béla Bartók" {
		t.Errorf("Composer = %q, want %q", tags["COMPOSER"], "Béla Bartók")
	}
	if tags["ARTIST"] != "Mstislav Rostropóvič" {
		t.Errorf("Artist = %q, want %q", tags["ARTIST"], "Mstislav Rostropóvič")
	}
	if tags["TITLE"] != "Sonate für Violine und Klavier: Frisch (Œuvre)" {
		t.Errorf("Title = %q, want %q", tags["TITLE"], "Sonate für Violine und Klavier: Frisch (Œuvre)")
	}
}

// TestFLACWriter_MultiplePerformers tests complex performer formatting.
func TestFLACWriter_MultiplePerformers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	sourcePath := generateTestFLAC(t, tmpDir, "source.flac")
	destPath := filepath.Join(tmpDir, "dest.flac")

	// Create metadata with multiple performers
	track := &domain.Track{
		File: domain.File{
			Path: "01.flac",
			Size: 0,
		},
		Disc:  1,
		Track: 1,
		Title: "Violin Concerto in D major, Op. 77: I. Allegro non troppo",
		Artists: []domain.Artist{
			{Name: "Johannes Brahms", Role: domain.RoleComposer},
			{Name: "Anne-Sophie Mutter", Role: domain.RoleSoloist},
			{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			{Name: "Herbert von Karajan", Role: domain.RoleConductor},
		},
	}
	torrent := &domain.Torrent{
		RootPath:     "brahms",
		Title:        "Brahms: Violin Concerto",
		OriginalYear: 1980,
		Edition: &domain.Edition{
			Label:         "Sony Classical",
			Year:          1992,
			CatalogNumber: "SK 52594",
		},
		Files: []domain.FileLike{track},
	}

	writer := NewFLACWriter()
	err := writer.WriteTrack(sourcePath, destPath, track, torrent)
	if err != nil {
		t.Fatalf("WriteTrack() error = %v", err)
	}

	// Verify all roles present
	tags := readVorbisComments(t, destPath)

	expected := map[string]string{
		"COMPOSER":  "Johannes Brahms",
		"ARTIST":    "Anne-Sophie Mutter, Berlin Philharmonic, Herbert von Karajan",
		"PERFORMER": "Anne-Sophie Mutter",
		"ENSEMBLE":  "Berlin Philharmonic",
		"CONDUCTOR": "Herbert von Karajan",
	}

	for key, want := range expected {
		if got := tags[key]; got != want {
			t.Errorf("Tag %q = %q, want %q", key, got, want)
		}
	}
}

// TestFLACWriter_NoEdition tests writing without edition information.
func TestFLACWriter_NoEdition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	sourcePath := generateTestFLAC(t, tmpDir, "source.flac")
	destPath := filepath.Join(tmpDir, "dest.flac")

	track := &domain.Track{
		File: domain.File{
			Path: "01.flac",
			Size: 0,
		},
		Disc:    1,
		Track:   1,
		Title:   "Test Track",
		Artists: []domain.Artist{{Name: "Test Composer", Role: domain.RoleComposer}},
	}
	torrent := &domain.Torrent{
		RootPath:     "test",
		Title:        "Test Album",
		OriginalYear: 2020,
		Edition:      nil,
		Files:        []domain.FileLike{track},
	}
	// No edition

	writer := NewFLACWriter()
	err := writer.WriteTrack(sourcePath, destPath, track, torrent)
	if err != nil {
		t.Fatalf("WriteTrack() error = %v", err)
	}

	tags := readVorbisComments(t, destPath)

	// Should have ORIGINALDATE but not DATE/LABEL/CATALOGNUMBER
	if tags["ORIGINALDATE"] != "2020" {
		t.Errorf("ORIGINALDATE = %q, want %q", tags["ORIGINALDATE"], "2020")
	}

	// These should not be present
	if _, exists := tags["DATE"]; exists {
		t.Error("DATE tag should not exist without edition")
	}
	if _, exists := tags["LABEL"]; exists {
		t.Error("LABEL tag should not exist without edition")
	}
	if _, exists := tags["CATALOGNUMBER"]; exists {
		t.Error("CATALOGNUMBER tag should not exist without edition")
	}
}

// TestFLACWriter_ErrorHandling tests error cases.
func TestFLACWriter_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewFLACWriter()
	track := &domain.Track{
		File: domain.File{
			Path: "01.flac",
			Size: 0,
		},
		Disc:    1,
		Track:   1,
		Title:   "Test",
		Artists: []domain.Artist{{Name: "Test", Role: domain.RoleComposer}},
	}

	torrent := &domain.Torrent{
		RootPath:     "test",
		Title:        "Test",
		OriginalYear: 2020,
		Edition: &domain.Edition{
			Label: "Sony Classical", Year: 1992, CatalogNumber: "SK 52594",
		},
		Files: []domain.FileLike{track},
	}

	tests := []struct {
		Name       string
		SourcePath string
		DestPath   string
		WantErr    bool
	}{
		{
			Name:       "source does not exist",
			SourcePath: filepath.Join(tmpDir, "nonexistent.flac"),
			DestPath:   filepath.Join(tmpDir, "dest.flac"),
			WantErr:    true,
		},
		{
			Name:       "source is not a FLAC file",
			SourcePath: filepath.Join(tmpDir, "notflac.txt"),
			DestPath:   filepath.Join(tmpDir, "dest.flac"),
			WantErr:    true,
		},
		{
			Name:       "dest directory does not exist",
			SourcePath: generateTestFLAC(t, tmpDir, "source.flac"),
			DestPath:   filepath.Join(tmpDir, "nonexistent", "dest.flac"),
			WantErr:    true,
		},
	}

	// Create the "not a FLAC" file
	os.WriteFile(filepath.Join(tmpDir, "notflac.txt"), []byte("not a flac"), 0644)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := writer.WriteTrack(tt.SourcePath, tt.DestPath, track, torrent)
			if (err != nil) != tt.WantErr {
				t.Errorf("WriteTrack() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}
