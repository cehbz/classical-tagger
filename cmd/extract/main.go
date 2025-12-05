package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

	localTorrent := extractFromDirectory(*dir)

	// Save local extraction
	localFile := baseName + ".json"
	if err := localTorrent.Save(localFile); err != nil {
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

	// get release(s)
	releases := []*discogs.Release{}
	if *releaseID != 0 {
		release, err := client.GetRelease(*releaseID)
		if err != nil || release == nil {
			fmt.Fprintf(os.Stderr, "Error fetching release: %v\n", err)
			os.Exit(1)
		}
		releases = append(releases, release)
	} else {
		// Search using extracted metadata
		artist := extractArtist(localTorrent)
		album := localTorrent.Title

		if artist == "" || album == "" {
			fmt.Fprintf(os.Stderr, "Warning: Cannot search Discogs without artist and album information\n")
			return
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "Searching Discogs for: artist=%q album=%q\n", artist, album)
		}

		releases, err = client.Search(artist, album)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Discogs search failed: %v\n", err)
			return
		}
		if len(releases) == 0 {
			fmt.Fprintf(os.Stderr, "No Discogs releases found for: %s - %s\n", artist, album)
			return
		}
	}

	// Handle search results
	if len(releases) > 1 {
		// Multiple matches - display and exit
		fmt.Fprintf(os.Stderr, "\nMultiple Discogs releases found:\n\n")

		releaseTemplate := `  [{{.ID}}] {{.Title}}{{if .Label}} - {{.Label}}{{end}}{{if .CatalogNumber}} {{.CatalogNumber}}{{end}}{{if gt .Year 0}} ({{.Year}}){{end}}{{if .Country}}, {{.Country}}{{end}}\n`
		tmpl := template.Must(template.New("release").Parse(releaseTemplate))
		for _, release := range releases {
			if err := tmpl.Execute(os.Stderr, release); err != nil {
				fmt.Fprintf(os.Stderr, "Error rendering template: %v\n", err)
			}
		}

		fmt.Fprintf(os.Stderr, "\nPlease re-run with --release-id to select a specific release:\n")
		fmt.Fprintf(os.Stderr, "  extract -dir %q --release-id XXXXXX\n\n", *dir)
		os.Exit(1)
	}

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
	// Use parent directory as rootPath so generated directory is a sibling of local directory
	parentDir := filepath.Dir(*dir)
	if err := release.SaveToFile(discogsFile, parentDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving Discogs data: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✓ Discogs metadata saved to: %s\n", discogsFile)
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
func extractFromDirectory(dirPath string) *domain.Torrent {
	album, err := scraping.ExtractFromDirectory(dirPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting from directory: %v\n", err)
		if !*force {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Forcing local extraction.\n")
		album = &domain.Album{
			Title: filepath.Base(dirPath),
		}
	}

	// Convert domain.Album to domain.Torrent
	torrent := album.ToTorrent(filepath.Base(dirPath))

	// Display extraction summary
	if torrent != nil {
		fmt.Fprintf(os.Stderr, "✓ Extracted: %s", torrent.Title)
		if torrent.OriginalYear > 0 {
			fmt.Fprintf(os.Stderr, " (%d)", torrent.OriginalYear)
		}
		fmt.Fprintf(os.Stderr, " - %d tracks\n", len(torrent.Tracks()))
	}

	return torrent
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
