package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cehbz/classical-tagger/internal/config"
	"github.com/cehbz/classical-tagger/internal/discogs"
	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/scraping"
)

var (
	dir        = flag.String("dir", "", "Directory containing FLAC files (required)")
	releaseID  = flag.Int("release-id", 0, "Specific Discogs release ID to use")
	outputFile = flag.String("output", "", "Base name for output files (default: directory name)")
	verbose    = flag.Bool("verbose", false, "Enable verbose output")
	force      = flag.Bool("force", false, "Create output even if required fields are missing")
	noAPI      = flag.Bool("no-api", false, "Skip Discogs API lookup")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	// Validate required arguments
	if *dir == "" {
		fmt.Fprintf(os.Stderr, "Error: -dir is required\n\n")
		usage()
		os.Exit(1)
	}

	// Verify directory exists
	if info, err := os.Stat(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Cannot access directory %s: %v\n", *dir, err)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", *dir)
		os.Exit(1)
	}

	// Determine output base name
	baseName := *outputFile
	if baseName == "" {
		baseName = filepath.Base(*dir)
		// Clean up the name (remove common suffixes)
		baseName = strings.TrimSuffix(baseName, " (FLAC)")
		baseName = strings.TrimSuffix(baseName, " FLAC")
	}

	// Step 1: Extract local metadata
	if *verbose {
		fmt.Fprintf(os.Stderr, "Extracting metadata from: %s\n", *dir)
	}

	localResult := extractFromDirectory(*dir)

	// Save local extraction
	localFile := baseName + ".json"
	if err := localResult.Torrent.Save(localFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving local metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✓ Local metadata saved to: %s\n", localFile)

	// Step 2: Try Discogs API (unless disabled)
	if *noAPI {
		if *verbose {
			fmt.Fprintf(os.Stderr, "Skipping Discogs API (--no-api specified)\n")
		}
		return
	}

	// Load Discogs token
	token, err := config.LoadDiscogsToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Cannot load Discogs token: %v\n", err)
		fmt.Fprintf(os.Stderr, "Continuing with local extraction only.\n")
		fmt.Fprintf(os.Stderr, "To enable Discogs lookup, create ~/.config/classical-tagger/config.yaml with your token.\n")
		return
	}

	client := discogs.NewClient(token)

	// If release ID provided, fetch directly
	if *releaseID != 0 {
		if *verbose {
			fmt.Fprintf(os.Stderr, "Fetching Discogs release #%d\n", *releaseID)
		}

		release, err := client.GetRelease(*releaseID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching release: %v\n", err)
			os.Exit(1)
		}

		discogsFile := baseName + "_discogs.json"
		if err := release.SaveToFile(discogsFile, baseName); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving Discogs data: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "✓ Discogs metadata saved to: %s\n", discogsFile)
		return
	}

	// Search using extracted metadata
	artist := extractArtist(localResult.Torrent)
	album := localResult.Torrent.Title

	if artist == "" || album == "" {
		fmt.Fprintf(os.Stderr, "Warning: Cannot search Discogs without artist and album information\n")
		return
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Searching Discogs for: artist=%q album=%q\n", artist, album)
	}

	releases, err := client.Search(artist, album)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Discogs search failed: %v\n", err)
		return
	}

	// Handle search results
	switch len(releases) {
	case 0:
		fmt.Fprintf(os.Stderr, "No Discogs releases found for: %s - %s\n", artist, album)

	case 1:
		// Single match - fetch automatically
		if *verbose {
			fmt.Fprintf(os.Stderr, "Found single match: %s - %s [%d]\n",
				releases[0].Label, releases[0].CatalogNumber, releases[0].ID)
		}

		release, err := client.GetRelease(releases[0].ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching release details: %v\n", err)
			os.Exit(1)
		}

		discogsFile := baseName + "_discogs.json"
		if err := release.SaveToFile(discogsFile, baseName); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving Discogs data: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "✓ Discogs metadata saved to: %s\n", discogsFile)

	default:
		// Multiple matches - display and exit
		fmt.Fprintf(os.Stderr, "\nMultiple Discogs releases found:\n\n")

		for i, release := range releases {
			fmt.Fprintf(os.Stderr, "  [%d] %s", release.ID, release.Title)

			if release.Label != "" {
				fmt.Fprintf(os.Stderr, " - %s", release.Label)
			}
			if release.CatalogNumber != "" {
				fmt.Fprintf(os.Stderr, " %s", release.CatalogNumber)
			}
			if release.Year > 0 {
				fmt.Fprintf(os.Stderr, " (%d)", release.Year)
			}
			if release.Country != "" {
				fmt.Fprintf(os.Stderr, ", %s", release.Country)
			}

			fmt.Fprintln(os.Stderr)

			// Limit display to first 10
			if i >= 9 && len(releases) > 10 {
				fmt.Fprintf(os.Stderr, "  ... and %d more\n", len(releases)-10)
				break
			}
		}

		fmt.Fprintf(os.Stderr, "\nPlease re-run with --release-id to select a specific release:\n")
		fmt.Fprintf(os.Stderr, "  extract -dir %q --release-id XXXXXX\n\n", *dir)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: extract -dir DIRECTORY [options]\n\n")
	fmt.Fprintf(os.Stderr, "Extract metadata from FLAC files and optionally enrich with Discogs data.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nOutput:\n")
	fmt.Fprintf(os.Stderr, "  Creates two files:\n")
	fmt.Fprintf(os.Stderr, "    <name>.json         - Metadata extracted from FLAC files\n")
	fmt.Fprintf(os.Stderr, "    <name>_discogs.json - Metadata from Discogs API (if available)\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Extract with automatic Discogs lookup:\n")
	fmt.Fprintf(os.Stderr, "  extract -dir \"/music/Bach - Goldberg Variations\"\n\n")
	fmt.Fprintf(os.Stderr, "  # Use specific Discogs release:\n")
	fmt.Fprintf(os.Stderr, "  extract -dir \"/music/Bach - Goldberg Variations\" --release-id 195873\n\n")
	fmt.Fprintf(os.Stderr, "  # Local extraction only:\n")
	fmt.Fprintf(os.Stderr, "  extract -dir \"/music/Bach - Goldberg Variations\" --no-api\n")
}

// extractFromDirectory extracts metadata from local FLAC files
func extractFromDirectory(dirPath string) *scraping.ExtractionResult {
	extractor := scraping.NewLocalExtractor()
	result, err := extractor.ExtractFromDirectory(dirPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting from directory: %v\n", err)
		os.Exit(1)
	}

	// Display extraction summary
	if result.Torrent != nil {
		fmt.Fprintf(os.Stderr, "✓ Extracted: %s", result.Torrent.Title)
		if result.Torrent.OriginalYear > 0 {
			fmt.Fprintf(os.Stderr, " (%d)", result.Torrent.OriginalYear)
		}
		fmt.Fprintf(os.Stderr, " - %d tracks\n", len(result.Torrent.Tracks()))
	}

	// Show warnings/errors
	for _, warning := range result.Warnings {
		fmt.Fprintf(os.Stderr, "  Warning: %s\n", warning)
	}

	for _, err := range result.Errors {
		if err.Required && !*force {
			fmt.Fprintf(os.Stderr, "  ERROR: %s - %s\n", err.Field, err.Message)
		} else {
			fmt.Fprintf(os.Stderr, "  Warning: %s - %s\n", err.Field, err.Message)
		}
	}

	// Fail if required errors and not forced
	hasRequiredErrors := false
	for _, err := range result.Errors {
		if err.Required {
			hasRequiredErrors = true
			break
		}
	}

	if hasRequiredErrors && !*force {
		fmt.Fprintf(os.Stderr, "\n❌ ERROR: Extraction failed due to required field errors\n")
		fmt.Fprintf(os.Stderr, "Use -force to create output anyway\n")
		os.Exit(1)
	}

	return result
}

// extractArtist attempts to get a searchable artist from the torrent
func extractArtist(t *domain.Torrent) string {
	if t == nil {
		return ""
	}

	// Try album artist first
	if len(t.AlbumArtist) > 0 {
		// Look for composer
		for _, artist := range t.AlbumArtist {
			if artist.Role == domain.RoleComposer {
				return artist.Name
			}
		}
		// Use first artist
		return t.AlbumArtist[0].Name
	}

	// Try to find composer from tracks
	tracks := t.Tracks()
	if len(tracks) > 0 && len(tracks[0].Artists) > 0 {
		for _, artist := range tracks[0].Artists {
			if artist.Role == domain.RoleComposer {
				return artist.Name
			}
		}
		// Use first artist from first track
		return tracks[0].Artists[0].Name
	}

	return ""
}
