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
	outputDir    = flag.String("output", "", "Output directory for tagged files (defaults to <targetDir>_tagged)")
	dryRun       = flag.Bool("dry-run", false, "Show what would be done without actually doing it")
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
	fmt.Printf("\nScanning directory: %s\n", *targetDir)
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

	// Determine output directory
	outDir := *outputDir
	if outDir == "" {
		outDir = *targetDir + "_tagged"
	}

	fmt.Println()

	// Apply tags
	if *dryRun {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Printf("Would write tagged files to: %s\n", outDir)
		fmt.Println("Would apply tags to the following files:")
		for track, file := range matches {
			if file != "" {
				destPath := filepath.Join(outDir, filepath.Base(file))
				fmt.Printf("  %s -> %s\n", filepath.Base(file), destPath)
				fmt.Printf("    Title: %s\n", track.Title())
				fmt.Printf("    Composer: %s\n", track.Composer().Name())
			}
		}
		fmt.Println("\nNo files were modified.")
		return
	}

	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Writing tagged files to: %s\n", outDir)
	writer := tagging.NewFLACWriter()

	successCount := 0
	errorCount := 0

	for track, file := range matches {
		if file == "" {
			continue
		}

		// Determine destination path
		destPath := filepath.Join(outDir, filepath.Base(file))

		// Write tags
		err := writer.WriteTrack(file, destPath, track, album)
		if err != nil {
			fmt.Printf("❌ Failed to write %s: %v\n", filepath.Base(file), err)
			errorCount++
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
	fmt.Printf("\n📁 Tagged files written to: %s\n", outDir)

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

// MatchTracksToFiles matches tracks to files based on track number in filename.
// Returns a map of track -> file path (empty string if no match found).
func MatchTracksToFiles(album *domain.Album, files []string) map[*domain.Track]string {
	matches := make(map[*domain.Track]string)

	for _, track := range album.Tracks() {
		matches[track] = ""

		// Try to find file by track number prefix
		trackPrefix := fmt.Sprintf("%02d", track.Track())

		for _, file := range files {
			base := filepath.Base(file)
			if strings.HasPrefix(base, trackPrefix) {
				matches[track] = file
				break
			}
		}
	}

	return matches
}
