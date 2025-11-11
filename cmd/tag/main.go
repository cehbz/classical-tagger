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
	torrent, err := LoadMetadataJSON(*metadataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Loaded torrent: %s (%d)\n", torrent.Title, torrent.OriginalYear)
	fmt.Printf("  Tracks: %d\n\n", len(torrent.Tracks()))

	// Validate metadata unless --force
	if !*force {
		fmt.Println("Validating metadata...")
		issues := validation.Check(torrent, nil)

		hasErrors := false
		for _, issue := range issues {
			switch issue.Level {
			case domain.LevelError:
				hasErrors = true
				fmt.Printf("❌ %s\n", issue)
			case domain.LevelWarning:
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
	matches := MatchTracksToFiles(torrent, files)

	unmatchedTracks := 0
	for track, file := range matches {
		if file == "" {
			unmatchedTracks++
			fmt.Printf("⚠️  No file found for track %d: %s\n", track.Track, track.Title)
		} else {
			fmt.Printf("✓ Track %d -> %s\n", track.Track, filepath.Base(file))
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
		// Use parent directory of targetDir as base, or current directory
		baseDir := filepath.Dir(*targetDir)
		if baseDir == "." || baseDir == *targetDir {
			baseDir = "."
		}
		// Generate directory name from torrent metadata
		dirName := torrent.DirectoryName()
		dir := filepath.Base(*targetDir)
		if dir == dirName {
			dirName = dirName + "_tagged"
		}
		outDir = filepath.Join(baseDir, dirName)
	}

	fmt.Println()

	// Check if multi-disc album
	isMultiDisc := torrent.IsMultiDisc()
	totalTracks := len(torrent.Tracks())

	// Apply tags
	if *dryRun {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Printf("Would write tagged files to: %s\n", outDir)
		if isMultiDisc {
			fmt.Println("Multi-disc album detected - will create disc subdirectories")
		}
		fmt.Println("Would apply tags to the following files:")
		for track, file := range matches {
			composers := track.Composers()
			composerName := ""
			if len(composers) > 0 {
				composerName = composers[0].Name
			}
			if file != "" {
				// Generate new filename
				newFilename := tagging.GenerateFilename(track, totalTracks)
				destPath := buildDestinationPath(outDir, track, newFilename, isMultiDisc)
				fmt.Printf("  %s -> %s\n", filepath.Base(file), destPath)
				fmt.Printf("    Title: %s\n", track.Title)
				fmt.Printf("    Composer: %s\n", composerName)
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
	if isMultiDisc {
		fmt.Println("Multi-disc album detected - creating disc subdirectories")
	}
	writer := tagging.NewFLACWriter()

	successCount := 0
	errorCount := 0

	for track, file := range matches {
		if file == "" {
			continue
		}

		// Generate new filename
		newFilename := tagging.GenerateFilename(track, totalTracks)
		destPath := buildDestinationPath(outDir, track, newFilename, isMultiDisc)

		// Create disc subdirectory if needed
		if isMultiDisc {
			discDir := filepath.Dir(destPath)
			if err := os.MkdirAll(discDir, 0755); err != nil {
				fmt.Printf("❌ Failed to create disc directory %s: %v\n", discDir, err)
				errorCount++
				continue
			}
		}

		// Write tags
		err := writer.WriteTrack(file, destPath, track, torrent)
		if err != nil {
			fmt.Printf("❌ Failed to write %s: %v\n", newFilename, err)
			errorCount++
			continue
		}

		fmt.Printf("✓ Created %s\n", destPath)
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

// LoadMetadataJSON loads torrent metadata from a JSON file.
func LoadMetadataJSON(path string) (*domain.Torrent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	repo := storage.NewRepository()
	torrent, err := repo.LoadFromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return torrent, nil
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
func MatchTracksToFiles(torrent *domain.Torrent, files []string) map[*domain.Track]string {
	matches := make(map[*domain.Track]string)

	for _, track := range torrent.Tracks() {
		matches[track] = ""

		// Try to find file by track number prefix
		trackPrefix := fmt.Sprintf("%02d", track.Track)

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

// buildDestinationPath builds the destination path for a track file.
// Handles multi-disc albums by creating subdirectories.
func buildDestinationPath(baseDir string, track *domain.Track, filename string, isMultiDisc bool) string {
	if isMultiDisc {
		// Create disc subdirectory for all discs in multi-disc albums
		discSubdir := tagging.GenerateDiscSubdirectoryName(track.Disc, "")
		return filepath.Join(baseDir, discSubdir, filename)
	}
	return filepath.Join(baseDir, filename)
}
