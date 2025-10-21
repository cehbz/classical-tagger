package tagging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// FLACWriter writes tags to FLAC files.
type FLACWriter struct {
	dryRun bool
}

// NewFLACWriter creates a new FLACWriter.
func NewFLACWriter() *FLACWriter {
	return &FLACWriter{
		dryRun: false,
	}
}

// SetDryRun enables or disables dry-run mode.
// In dry-run mode, no files are actually modified.
func (w *FLACWriter) SetDryRun(dryRun bool) {
	w.dryRun = dryRun
}

// WriteTrack writes track metadata to a FLAC file.
func (w *FLACWriter) WriteTrack(path string, track *domain.Track) error {
	if w.dryRun {
		fmt.Printf("[DRY RUN] Would write track: %s\n", path)
		return nil
	}
	
	// For now, return not implemented
	// Full implementation requires a FLAC tag writing library
	// The dhowden/tag library only supports reading
	return fmt.Errorf("FLAC writing not yet implemented - requires tag writing library")
}

// WriteAlbum writes all tracks in an album to their respective files.
// Returns a map of file paths to any errors encountered.
func (w *FLACWriter) WriteAlbum(basePath string, album *domain.Album) map[string]error {
	errors := make(map[string]error)
	
	for _, track := range album.Tracks() {
		// Construct file path
		var filePath string
		if track.Name() != "" {
			filePath = filepath.Join(basePath, track.Name())
		} else {
			// Construct default filename
			filename := fmt.Sprintf("%02d %s.flac", track.Track(), track.Title())
			filePath = filepath.Join(basePath, filename)
		}
		
		err := w.WriteTrack(filePath, track)
		if err != nil {
			errors[filePath] = err
		}
	}
	
	return errors
}

// BackupFile creates a backup copy of a file with a .bak extension and timestamp.
func (w *FLACWriter) BackupFile(path string) (string, error) {
	if w.dryRun {
		backupPath := path + ".bak"
		fmt.Printf("[DRY RUN] Would backup: %s -> %s\n", path, backupPath)
		return backupPath, nil
	}
	
	// Read original file
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	
	// Create backup path with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.bak", path, timestamp)
	
	// Write backup
	err = os.WriteFile(backupPath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}
	
	return backupPath, nil
}

// RestoreBackup restores a file from a backup.
func (w *FLACWriter) RestoreBackup(backupPath, originalPath string) error {
	if w.dryRun {
		fmt.Printf("[DRY RUN] Would restore: %s -> %s\n", backupPath, originalPath)
		return nil
	}
	
	// Read backup
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}
	
	// Write to original location
	err = os.WriteFile(originalPath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}
	
	return nil
}

// TrackToMetadata converts a domain Track to a Metadata struct for writing.
// This is the inverse of Metadata.ToTrack().
func TrackToMetadata(track *domain.Track, albumTitle string, albumYear int) Metadata {
	// Get composer
	composer := track.Composer()
	
	// Format artists as "Soloist, Ensemble, Conductor"
	var artistParts []string
	for _, artist := range track.Artists() {
		// Skip composer (already in Composer field)
		if artist.Role() == domain.RoleComposer {
			continue
		}
		artistParts = append(artistParts, artist.Name())
	}
	artistString := strings.Join(artistParts, ", ")
	
	return Metadata{
		Title:       track.Title(),
		Artist:      artistString,
		Album:       albumTitle,
		Composer:    composer.Name(),
		Year:        strconv.Itoa(albumYear),
		TrackNumber: strconv.Itoa(track.Track()),
		DiscNumber:  strconv.Itoa(track.Disc()),
	}
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}
	
	return destFile.Sync()
}

// Note: Full FLAC writing implementation requires additional library
// The dhowden/tag library only supports reading tags, not writing them.
// 
// For production use, we would need to either:
// 1. Use a different library like go-flac or flac
// 2. Shell out to metaflac command-line tool
// 3. Implement FLAC vorbis comment writing from scratch
//
// For now, we provide the interface and dry-run mode for testing.
