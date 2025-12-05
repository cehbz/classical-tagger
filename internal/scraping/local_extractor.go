package scraping

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/dhowden/tag"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
)

// ExtractFromDirectory reads all FLAC files in a directory and extracts metadata.
// It attempts to build a complete domain.Album structure from the tags and filenames.
func ExtractFromDirectory(dirPath string) (*domain.Album, error) {
	// Verify directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("directory access error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Find all FLAC files
	flacFiles, err := findFLACFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error finding FLAC files: %w", err)
	}

	if len(flacFiles) == 0 {
		return nil, fmt.Errorf("no FLAC files found in directory")
	}

	// Extract metadata from files
	return extractFromFiles(flacFiles, dirPath)
}

// findFLACFiles recursively finds all FLAC files in a directory.
func findFLACFiles(dirPath string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".flac") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files by path for consistent ordering
	sort.Strings(files)

	return files, nil
}

// extractFromFiles extracts metadata from a list of FLAC files.
func extractFromFiles(files []string, dirPath string) (*domain.Album, error) {
	// Create initial album data with sentinel values
	album := &domain.Album{
		FolderName:   filepath.Base(dirPath),
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]*domain.Track, 0, len(files)),
	}

	// Extract album-level metadata from first file
	if len(files) > 0 {
		albumData, warning := extractAlbumMetadata(files[0])
		album.Title = albumData.Title
		album.OriginalYear = albumData.OriginalYear
		album.Edition = albumData.Edition
		album.AlbumArtist = albumData.AlbumArtist

		if warning != "" {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
		}
	}

	// Extract track metadata from each file and collect ALBUMARTIST values
	trackAlbumArtists := make(map[string]bool) // Track unique ALBUMARTIST values
	for _, filePath := range files {
		track, albumArtistValue, err := extractTrackMetadataWithAlbumArtist(filePath, dirPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: file %s: %v\n", filepath.Base(filePath), err)
			continue
		}

		// Track ALBUMARTIST value for verification
		if albumArtistValue != "" {
			trackAlbumArtists[albumArtistValue] = true
		}

		album.Tracks = append(album.Tracks, track)
	}

	// Validate we got tracks
	if len(album.Tracks) == 0 {
		return nil, fmt.Errorf("no tracks extracted")
	}

	// Verify ALBUMARTIST consistency across tracks
	if len(trackAlbumArtists) > 1 {
		// Multiple different ALBUMARTIST values found
		fmt.Fprintf(os.Stderr, "Warning: inconsistent ALBUMARTIST tags across tracks: %v\n", trackAlbumArtists)
	} else if len(trackAlbumArtists) == 1 {
		// All tracks have the same ALBUMARTIST string
		trackAlbumArtistStr := ""
		for aa := range trackAlbumArtists {
			trackAlbumArtistStr = aa
			break
		}

		if len(album.AlbumArtist) == 0 {
			// Album-level ALBUMARTIST not set, but tracks have it - parse from track value
			album.AlbumArtist = domain.ParseArtistField(trackAlbumArtistStr)
		} else {
			// Compare formatted strings
			albumArtistStr := domain.FormatArtists(album.AlbumArtist)
			if albumArtistStr != trackAlbumArtistStr {
				// Album-level and track-level ALBUMARTIST differ
				fmt.Fprintf(os.Stderr, "Warning: album-level ALBUMARTIST '%s' differs from track-level '%s'\n", albumArtistStr, trackAlbumArtistStr)
			}
		}
	}

	// If album artist is already set (from tags), refine roles using universal performers from tracks
	// This ensures we have accurate roles based on actual track performers
	if len(album.AlbumArtist) > 0 && len(album.Tracks) > 0 {
		universalArtists := album.AlbumArtists()
		if len(universalArtists) > 0 {
			// Use universal performers (they have correct roles from tracks)
			// Compare names to ensure they match what we parsed from tags
			parsedNames := make(map[string]bool)
			for _, artist := range album.AlbumArtist {
				parsedNames[artist.Name] = true
			}
			universalNames := make(map[string]bool)
			for _, artist := range universalArtists {
				universalNames[artist.Name] = true
			}

			// If names match, use universal artists (better roles)
			namesMatch := len(parsedNames) == len(universalNames)
			if namesMatch {
				for name := range parsedNames {
					if !universalNames[name] {
						namesMatch = false
						break
					}
				}
			}

			if namesMatch {
				album.AlbumArtist = universalArtists
			}

			// Ensure AlbumArtist performers are present on each track (unless Various Artists)
			if !strings.EqualFold(strings.TrimSpace(domain.FormatArtists(album.AlbumArtist)), "Various Artists") {
				ensureArtistsOnTracks(album.Tracks, album.AlbumArtist)
			}
		}
	}

	// If album artist is empty but we have performers in all tracks, synthesize it
	if len(album.AlbumArtist) == 0 && len(album.Tracks) > 0 {
		universalArtists := album.AlbumArtists()
		if len(universalArtists) > 0 {
			album.AlbumArtist = universalArtists
			// Ensure AlbumArtist performers are present on each track (unless Various Artists)
			if !strings.EqualFold(strings.TrimSpace(domain.FormatArtists(album.AlbumArtist)), "Various Artists") {
				ensureArtistsOnTracks(album.Tracks, album.AlbumArtist)
			}
		}
	}

	// Try to extract folder name metadata if album title missing
	if album.Title == MissingTitle {
		if _, title, year := parseDirectoryName(dirPath); title != "" {
			album.Title = title
			if year > 0 && album.OriginalYear == MissingYear {
				album.OriginalYear = year
			}
			fmt.Fprintf(os.Stderr, "Warning: album title extracted from directory name\n")
		}
	}

	// Check for missing required fields
	if album.Title == MissingTitle {
		fmt.Fprintf(os.Stderr, "Warning: title not found in tags or directory name\n")
	}
	if album.OriginalYear == MissingYear {
		fmt.Fprintf(os.Stderr, "Warning: year not found in tags or directory name\n")
	}

	return album, nil
}

// albumMetadata is a temporary structure for album-level data
type albumMetadata struct {
	Title        string
	OriginalYear int
	Edition      *domain.Edition
	AlbumArtist  []domain.Artist
}

// extractAlbumMetadata extracts album-level metadata from a FLAC file's tags.
// Returns the metadata and an optional warning string.
func extractAlbumMetadata(filePath string) (albumMetadata, string) {
	meta := albumMetadata{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		AlbumArtist:  nil,
	}

	f, err := os.Open(filePath)
	if err != nil {
		return meta, fmt.Sprintf("failed to open file for album Metadata: %v", err)
	}
	defer f.Close()

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		return meta, fmt.Sprintf("failed to read tags for album Metadata: %v", err)
	}

	// Extract album title
	if album := metadata.Album(); album != "" {
		meta.Title = album
	}

	// Extract year - prioritize ORIGINALDATE tag, fall back to standard YEAR tag
	vorbisTags := readVorbisCommentTags(filePath)

	// Check for DJ tags - error exit if found
	if djTag := vorbisTags["DJ"]; djTag != "" {
		fmt.Fprintf(os.Stderr, "Error: DJ tag detected in album metadata: %s. DJ tags are not yet supported.\n", filePath)
		os.Exit(1)
	}

	if originalDate := vorbisTags["ORIGINALDATE"]; originalDate != "" {
		if year, err := strconv.Atoi(originalDate); err == nil && year > 0 {
			meta.OriginalYear = year
		}
	} else if year := metadata.Year(); year > 0 {
		meta.OriginalYear = year
	}

	// Extract album artist
	if albumArtistStr := metadata.AlbumArtist(); albumArtistStr != "" {
		// Parse the string into artists (roles will be inferred)
		meta.AlbumArtist = domain.ParseArtistField(albumArtistStr)
	}

	// Extract edition info - prioritize direct tags, fall back to COMMENT parsing
	edition := extractEditionFromTags(vorbisTags)
	if edition == nil {
		// Fall back to COMMENT field parsing
		if comment := metadata.Comment(); comment != "" {
			edition = extractEditionFromComment(comment)
		}
	}
	if edition != nil {
		meta.Edition = edition
	}

	return meta, ""
}

// readVorbisCommentTags reads all Vorbis comment tags from a FLAC file.
// Returns a map of tag names (uppercase) to values.
func readVorbisCommentTags(filePath string) map[string]string {
	tags := make(map[string]string)

	flacFile, err := flac.ParseFile(filePath)
	if err != nil {
		return tags
	}

	// Find VorbisComment block
	for _, metaBlock := range flacFile.Meta {
		if metaBlock.Type == flac.VorbisComment {
			cmtBlock, err := flacvorbis.ParseFromMetaDataBlock(*metaBlock)
			if err != nil {
				continue
			}

			// Extract all comments
			for _, comment := range cmtBlock.Comments {
				parts := strings.SplitN(comment, "=", 2)
				if len(parts) == 2 {
					tagName := strings.ToUpper(parts[0])
					tagValue := parts[1]
					tags[tagName] = tagValue
				}
			}
			break
		}
	}

	return tags
}

// extractEditionFromTags extracts edition information from Vorbis comment tags.
// Returns nil if no edition data found.
func extractEditionFromTags(tags map[string]string) *domain.Edition {
	edition := &domain.Edition{}
	found := false

	// Read LABEL tag
	if label := tags["LABEL"]; label != "" {
		edition.Label = strings.TrimSpace(label)
		found = true
	}

	// Read CATALOGNUMBER tag
	if catalog := tags["CATALOGNUMBER"]; catalog != "" {
		edition.CatalogNumber = strings.TrimSpace(catalog)
		found = true
	}

	// Read DATE tag (edition year)
	if dateStr := tags["DATE"]; dateStr != "" {
		if year, err := strconv.Atoi(strings.TrimSpace(dateStr)); err == nil && year > 0 {
			edition.Year = year
			found = true
		}
	}

	if !found {
		return nil
	}

	return edition
}

// extractEditionFromComment attempts to extract label/catalog from comment field.
// Returns nil if no edition data found.
func extractEditionFromComment(comment string) *domain.Edition {
	edition := &domain.Edition{}
	found := false

	// Look for patterns like "Label: Deutsche Grammophon" or "Catalog: 479 1234"
	labelPattern := regexp.MustCompile(`(?i)label:\s*(.+?)(?:\n|$)`)
	catalogPattern := regexp.MustCompile(`(?i)catalog(?:\s*number)?:\s*([A-Z0-9\-\s]+)`)

	if matches := labelPattern.FindStringSubmatch(comment); len(matches) > 1 {
		edition.Label = strings.TrimSpace(matches[1])
		found = true
	}

	if matches := catalogPattern.FindStringSubmatch(comment); len(matches) > 1 {
		edition.CatalogNumber = strings.TrimSpace(matches[1])
		found = true
	}

	if !found {
		return nil
	}

	return edition
}

// extractTrackMetadataWithAlbumArtist extracts track-level metadata and also returns ALBUMARTIST value.
func extractTrackMetadataWithAlbumArtist(filePath string, baseDir string) (*domain.Track, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read tags: %w", err)
	}

	track := &domain.Track{
		File: domain.File{
			Path: "", // Will be set below
		},
		Disc:    1, // Default
		Track:   0,
		Title:   "",
		Artists: make([]domain.Artist, 0),
	}

	// Extract track number
	trackNum, _ := metadata.Track()
	if trackNum > 0 {
		track.Track = trackNum
	} else {
		// Try to extract from filename
		if num := extractTrackNumberFromFilename(filePath); num > 0 {
			track.Track = num
		} else {
			return track, "", fmt.Errorf("no track number found in tags or filename")
		}
	}

	// Extract disc number
	discNum, _ := metadata.Disc()
	if discNum > 0 {
		track.Disc = discNum
	} else {
		// Try to extract from path (CD1, CD2, etc.)
		track.Disc = extractDiscFromPath(filePath)
	}

	// Extract title
	if title := metadata.Title(); title != "" {
		track.Title = title
	} else {
		// Use filename without extension as fallback
		track.Title = extractTitleFromFilename(filePath)
	}

	// Extract composer (required field)
	if composer := metadata.Composer(); composer != "" {
		track.Artists = append(track.Artists, domain.Artist{Name: composer, Role: domain.RoleComposer})
	} else {
		return track, "", fmt.Errorf("no composer found in tags")
	}

	// Extract artists
	if artist := metadata.Artist(); artist != "" {
		track.Artists = append(track.Artists, domain.ParseArtistField(artist)...)
	} else if albumArtist := metadata.AlbumArtist(); albumArtist != "" {
		// Fallback to album artist if artist tag missing
		track.Artists = append(track.Artists, domain.ParseArtistField(albumArtist)...)
	}

	// Extract ALBUMARTIST value for verification (but don't store in track)
	albumArtistValue := metadata.AlbumArtist()

	// Check for DJ tags - error exit if found
	vorbisTags := readVorbisCommentTags(filePath)
	if djTag := vorbisTags["DJ"]; djTag != "" {
		fmt.Fprintf(os.Stderr, "Error: DJ tag detected in file: %s. DJ tags are not yet supported.\n", filePath)
		os.Exit(1)
	}

	// Set relative filename (add before the final return)
	relPath, err := filepath.Rel(baseDir, filePath)
	if err == nil {
		// Convert to forward slashes for consistency
		track.File.Path = filepath.ToSlash(relPath)
	} else {
		track.File.Path = filepath.Base(filePath)
	}

	return track, albumArtistValue, nil
}

// extractTrackNumberFromFilename attempts to extract track number from filename.
// Supports formats: "01 Title.flac", "01-Title.flac", "01.Title.flac", "01_Title.flac"
func extractTrackNumberFromFilename(filePath string) int {
	filename := filepath.Base(filePath)

	// Pattern: starts with 1-3 digits followed by separator
	pattern := regexp.MustCompile(`^(\d{1,3})[\s\-._]`)
	matches := pattern.FindStringSubmatch(filename)

	if len(matches) > 1 {
		num, err := strconv.Atoi(matches[1])
		if err == nil && num > 0 && num < 1000 {
			return num
		}
	}

	return 0
}

// extractDiscFromPath attempts to extract disc number from file path.
// Looks for "CD1", "CD2", "Disc 1", "Disc 2", etc.
func extractDiscFromPath(filePath string) int {
	// Check directory names in path
	parts := strings.Split(filepath.Dir(filePath), string(filepath.Separator))

	for _, part := range parts {
		// Pattern: CD1, CD2, Disc 1, Disc 2, etc.
		pattern := regexp.MustCompile(`(?i)(?:CD|Disc|Disk)\s*(\d+)`)
		matches := pattern.FindStringSubmatch(part)

		if len(matches) > 1 {
			num, err := strconv.Atoi(matches[1])
			if err == nil && num > 0 && num < 100 {
				return num
			}
		}
	}

	return 1 // Default to disc 1
}

// extractTitleFromFilename extracts title from filename as fallback.
// Removes track number prefix and file extension.
func extractTitleFromFilename(filePath string) string {
	filename := filepath.Base(filePath)

	// Remove extension
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	// Remove track number prefix (01, 01-, 01., 01_, etc.)
	pattern := regexp.MustCompile(`^\d{1,3}[\s\-._]*`)
	filename = pattern.ReplaceAllString(filename, "")

	// Clean up
	filename = strings.TrimSpace(filename)

	return filename
}

// parseDirectoryName attempts to extract album title and year from directory name.
// Handles formats like "Beethoven - Symphony No. 5 [1963]" or "Bach - Goldberg Variations (1741)"
func parseDirectoryName(dirPath string) (folderName string, title string, year int) {
	dirName := filepath.Base(dirPath)

	// Extract year from brackets or parentheses
	yearPattern := regexp.MustCompile(`[\[\(](\d{4})[\]\)]`)
	if matches := yearPattern.FindStringSubmatch(dirName); len(matches) > 1 {
		year, _ = strconv.Atoi(matches[1])
	}

	// Remove year and format indicators for title
	title = yearPattern.ReplaceAllString(dirName, "")

	// Remove format indicators like [FLAC], [MP3], etc.
	formatPattern := regexp.MustCompile(`\s*\[(FLAC|MP3|AAC|ALAC|WAV|APE|WV|24-\d+|16-\d+)\]`)
	title = formatPattern.ReplaceAllString(title, "")

	// Clean up whitespace
	title = strings.TrimSpace(title)

	return dirName, title, year
}
