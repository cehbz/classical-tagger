package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/storage"
	"github.com/cehbz/classical-tagger/internal/tagging"
	"github.com/cehbz/classical-tagger/internal/validation"
)

var (
	metadataFile = flag.String("metadata", "", "Path to metadata JSON file (required)")
	targetDir    = flag.String("dir", ".", "Target directory containing FLAC files")
	dryRun       = flag.Bool("dry-run", false, "Show what would be done without actually doing it")
	backup       = flag.Bool("backup", true, "Create backup before modifying files")
	force        = flag.Bool("force", false, "Skip validation and apply tags anyway")
)

func main() {
	flag.Parse()

	if *metadataFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -metadata flag is required\n")
		fmt.Fprintf(os.Stderr, "\nUsage: tag -metadata FILE [options]\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load metadata JSON
	fmt.Printf("Loading metadata from %s...\n", *metadataFile)
	album, err := LoadMetadataJSON(*metadataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Loaded album: %s (%d)\n", album.Title(), album.OriginalYear())
	fmt.Printf("  Tracks: %d\n\n", len(album.Tracks()))

	// Validate metadata unless --force
	if !*force {
		fmt.Println("Validating metadata...")
		validator := validation.NewAlbumValidator()
		issues := validator.ValidateMetadata(album)

		hasErrors := false
		for _, issue := range issues {
			if issue.Level() == domain.LevelError {
				hasErrors = true
				fmt.Printf("❌ %s\n", issue)
			} else if issue.Level() == domain.LevelWarning {
				fmt.Printf("⚠️  %s\n", issue)
			}
		}

		if hasErrors {
			fmt.Fprintf(os.Stderr, "\n❌ Metadata has errors. Fix them or use --force to proceed anyway.\n")
			os.Exit(1)
		}

		if len(issues) == 0 {
			fmt.Println("✓ Metadata is valid")
		} else {
			fmt.Println("⚠️  Metadata has warnings but is usable")
		}
	}

	// Find FLAC files in target directory
	fmt.Printf("Scanning directory: %s\n", *targetDir)
	files, err := FindFLACFiles(*targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Found %d FLAC files\n\n", len(files))

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No FLAC files found in directory\n")
		os.Exit(1)
	}

	// Match tracks to files
	fmt.Println("Matching tracks to files...")
	matches := MatchTracksToFiles(album, files)

	unmatchedTracks := 0
	for track, file := range matches {
		if file == "" {
			unmatchedTracks++
			fmt.Printf("⚠️  No file found for track %d: %s\n", track.Track(), track.Title())
		} else {
			fmt.Printf("✓ Track %d -> %s\n", track.Track(), filepath.Base(file))
		}
	}

	if unmatchedTracks > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  %d tracks could not be matched to files\n", unmatchedTracks)
		if !*force {
			fmt.Fprintf(os.Stderr, "Use --force to proceed anyway\n")
			os.Exit(1)
		}
	}

	fmt.Println()

	// Apply tags
	if *dryRun {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Println("Would apply tags to the following files:")
		for track, file := range matches {
			if file != "" {
				fmt.Printf("  %s\n", file)
				fmt.Printf("    Title: %s\n", track.Title())
				fmt.Printf("    Composer: %s\n", track.Composer().Name())
			}
		}
		fmt.Println("\nNo files were modified.")
		return
	}

	fmt.Println("Applying tags...")
	writer := tagging.NewFLACWriter()
	writer.SetDryRun(*dryRun)

	successCount := 0
	errorCount := 0
	backupPaths := make(map[string]string)

	for track, file := range matches {
		if file == "" {
			continue
		}

		// Backup if requested
		if *backup {
			backupPath, err := writer.BackupFile(file)
			if err != nil {
				fmt.Printf("❌ Failed to backup %s: %v\n", file, err)
				errorCount++
				continue
			}
			backupPaths[file] = backupPath
		}

		// Write tags
		err := writer.WriteTrack(file, track, album)
		if err != nil {
			fmt.Printf("❌ Failed to write tags to %s: %v\n", file, err)
			errorCount++

			// Restore backup if write failed
			if *backup {
				if backupPath, ok := backupPaths[file]; ok {
					writer.RestoreBackup(backupPath, file)
					fmt.Printf("   Restored from backup\n")
				}
			}
			continue
		}

		fmt.Printf("✓ Updated %s\n", filepath.Base(file))
		successCount++
	}

	// Summary
	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("✓ Successfully updated: %d files\n", successCount)
	if errorCount > 0 {
		fmt.Printf("❌ Errors: %d files\n", errorCount)
	}

	if *backup && successCount > 0 {
		fmt.Printf("\n💾 Backups created:\n")
		for _, backupPath := range backupPaths {
			fmt.Printf("  %s\n", backupPath)
		}
	}

	if errorCount > 0 {
		os.Exit(1)
	}
}

// LoadMetadataJSON loads album metadata from a JSON file.
func LoadMetadataJSON(path string) (*domain.Album, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	repo := storage.NewRepository()
	album, err := repo.LoadFromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return album, nil
}

// FindFLACFiles recursively finds all FLAC files in a directory.
func FindFLACFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".flac") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// MatchTracksToFiles matches tracks to files based on filename.
func MatchTracksToFiles(album *domain.Album, files []string) map[*domain.Track]string {
	matches := make(map[*domain.Track]string)

	// Create a map of basenames to full paths for quick lookup
	fileMap := make(map[string]string)
	for _, file := range files {
		basename := filepath.Base(file)
		fileMap[strings.ToLower(basename)] = file
	}

	for _, track := range album.Tracks() {
		if track.Name() != "" {
			// Try exact match first
			if file, ok := fileMap[strings.ToLower(track.Name())]; ok {
				matches[track] = file
				continue
			}

			// Try fuzzy match
			matched, file := MatchTrackToFile(track.Name(), files)
			if matched {
				matches[track] = file
				continue
			}
		}

		// No match found
		matches[track] = ""
	}

	return matches
}

// MatchTrackToFile attempts to match a track name to a file.
// Returns true and the matched file if found.
func MatchTrackToFile(trackName string, files []string) (bool, string) {
	trackLower := strings.ToLower(trackName)

	for _, file := range files {
		basename := filepath.Base(file)
		if strings.ToLower(basename) == trackLower {
			return true, file
		}
	}

	return false, ""
}
