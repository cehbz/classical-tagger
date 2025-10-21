package tagging

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestFLACWriter_WriteTrack(t *testing.T) {
	t.Skip("Requires FLAC test fixture and write permissions")
	
	writer := NewFLACWriter()
	
	// Create test track
	composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	performer, _ := domain.NewArtist("Glenn Gould", domain.RoleSoloist)
	track, _ := domain.NewTrack(1, 1, "Aria", []domain.Artist{composer, performer})
	track = track.WithName("01 Aria.flac")
	
	// This would write to a temp file in real test
	tempFile := filepath.Join(t.TempDir(), "test.flac")
	
	err := writer.WriteTrack(tempFile, track)
	if err != nil {
		t.Fatalf("WriteTrack() error = %v", err)
	}
	
	// Verify the tags were written
	reader := NewFLACReader()
	metadata, err := reader.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	
	if metadata.Composer != "Johann Sebastian Bach" {
		t.Errorf("Composer = %v, want 'Johann Sebastian Bach'", metadata.Composer)
	}
}

func TestFLACWriter_BackupFile(t *testing.T) {
	writer := NewFLACWriter()
	
	// Create temp file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.flac")
	content := []byte("test content")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Backup
	backupPath, err := writer.BackupFile(testFile)
	if err != nil {
		t.Fatalf("BackupFile() error = %v", err)
	}
	
	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
	
	// Verify backup content matches
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	
	if string(backupContent) != string(content) {
		t.Error("Backup content does not match original")
	}
}

func TestFLACWriter_RestoreBackup(t *testing.T) {
	writer := NewFLACWriter()
	
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.flac")
	originalContent := []byte("original content")
	
	// Create original file
	err := os.WriteFile(testFile, originalContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Backup
	backupPath, err := writer.BackupFile(testFile)
	if err != nil {
		t.Fatalf("BackupFile() error = %v", err)
	}
	
	// Modify original
	err = os.WriteFile(testFile, []byte("modified content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}
	
	// Restore
	err = writer.RestoreBackup(backupPath, testFile)
	if err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}
	
	// Verify restored content
	restoredContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	
	if string(restoredContent) != string(originalContent) {
		t.Error("Restored content does not match original")
	}
}

func TestFLACWriter_DryRun(t *testing.T) {
	writer := NewFLACWriter()
	writer.SetDryRun(true)
	
	// Create test track
	composer, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
	track, _ := domain.NewTrack(1, 1, "Symphony No. 1", []domain.Artist{composer})
	
	// In dry-run mode, should not fail even with non-existent file
	err := writer.WriteTrack("/nonexistent/path.flac", track)
	
	// Should not error in dry-run mode
	if err != nil {
		t.Errorf("DryRun mode should not error, got: %v", err)
	}
}

func TestTrackToMetadata(t *testing.T) {
	composer, _ := domain.NewArtist("Ludwig van Beethoven", domain.RoleComposer)
	soloist, _ := domain.NewArtist("Martha Argerich", domain.RoleSoloist)
	ensemble, _ := domain.NewArtist("London Symphony Orchestra", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Claudio Abbado", domain.RoleConductor)
	
	track, _ := domain.NewTrack(
		2, 
		5, 
		"Piano Concerto No. 1 in C major, Op. 15: III. Rondo",
		[]domain.Artist{composer, soloist, ensemble, conductor},
	)
	
	metadata := TrackToMetadata(track, "Test Album", 1970)
	
	if metadata.Composer != "Ludwig van Beethoven" {
		t.Errorf("Composer = %v, want 'Ludwig van Beethoven'", metadata.Composer)
	}
	
	if metadata.Title != "Piano Concerto No. 1 in C major, Op. 15: III. Rondo" {
		t.Errorf("Title = %v, want track title", metadata.Title)
	}
	
	// Artist should be formatted as "Soloist, Ensemble, Conductor"
	expectedArtist := "Martha Argerich, London Symphony Orchestra, Claudio Abbado"
	if metadata.Artist != expectedArtist {
		t.Errorf("Artist = %v, want %v", metadata.Artist, expectedArtist)
	}
	
	if metadata.TrackNumber != "5" {
		t.Errorf("TrackNumber = %v, want '5'", metadata.TrackNumber)
	}
	
	if metadata.DiscNumber != "2" {
		t.Errorf("DiscNumber = %v, want '2'", metadata.DiscNumber)
	}
}