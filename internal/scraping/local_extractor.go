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

// LocalExtractor extracts metadata from existing FLAC files in a directory.
// This is useful for converting already-tagged albums to our JSON format.
// It is immutable and stateless.
type LocalExtractor struct {
	// Extractor is stateless
}

// NewLocalExtractor creates a new local extractor.
func NewLocalExtractor() *LocalExtractor {
	return &LocalExtractor{}
}

// ExtractFromDirectory reads all FLAC files in a directory and extracts metadata.
// It attempts to build a complete domain.Album structure from the tags and filenames.
// Returns an immutable ExtractionResult.
func (e *LocalExtractor) ExtractFromDirectory(dirPath string) (*ExtractionResult, error) {
	// Verify directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("directory access error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Find all FLAC files
	flacFiles, err := e.findFLACFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error finding FLAC files: %w", err)
	}

	if len(flacFiles) == 0 {
		return nil, fmt.Errorf("no FLAC files found in directory")
	}

	// Extract metadata from files
	return e.extractFromFiles(flacFiles, dirPath)
}

// findFLACFiles recursively finds all FLAC files in a directory.
func (e *LocalExtractor) findFLACFiles(dirPath string) ([]string, error) {
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
func (e *LocalExtractor) extractFromFiles(files []string, dirPath string) (*ExtractionResult, error) {
	// Create initial album data with sentinel values
	data := &domain.Album{
		FolderName:   filepath.Base(dirPath),
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]*domain.Track, 0, len(files)),
	}

	// Start with empty result (will convert Album to Torrent at end)
	result := &ExtractionResult{
		Source: "local_directory",
	}
	parsingNotes := make(map[string]interface{})
	parsingNotes["source"] = "local_directory"
	parsingNotes["directory"] = dirPath
	parsingNotes["file_count"] = len(files)

	// Extract album-level metadata from first file
	if len(files) > 0 {
		albumData, warning := e.extractAlbumMetadata(files[0])
		data.Title = albumData.Title
		data.OriginalYear = albumData.OriginalYear
		data.Edition = albumData.Edition
		data.AlbumArtist = albumData.AlbumArtist

		if warning != "" {
			result.Warnings = append(result.Warnings, warning)
		}
	}

	// Extract track metadata from each file and collect ALBUMARTIST values
	trackAlbumArtists := make(map[string]bool) // Track unique ALBUMARTIST values
	for _, filePath := range files {
		track, albumArtistValue, err := e.extractTrackMetadataWithAlbumArtist(filePath, dirPath)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("file %s: %v", filepath.Base(filePath), err))
			continue
		}

		// Track ALBUMARTIST value for verification
		if albumArtistValue != "" {
			trackAlbumArtists[albumArtistValue] = true
		}

		data.Tracks = append(data.Tracks, track)
	}

	// Validate we got tracks
	if len(data.Tracks) == 0 {
		result.Warnings = append(result.Warnings, "no tracks extracted")
		// Convert Album to Torrent before returning
		result.Torrent = data.ToTorrent(filepath.Base(dirPath))
		return result, nil
	}

	// Verify ALBUMARTIST consistency across tracks
	if len(trackAlbumArtists) > 1 {
		// Multiple different ALBUMARTIST values found
		result.Errors = append(result.Errors, ExtractionError{
			Field:    "album_artist",
			Message:  fmt.Sprintf("inconsistent ALBUMARTIST tags across tracks: %v", trackAlbumArtists),
			Required: false,
		})
	} else if len(trackAlbumArtists) == 1 {
		// All tracks have the same ALBUMARTIST string
		trackAlbumArtistStr := ""
		for aa := range trackAlbumArtists {
			trackAlbumArtistStr = aa
			break
		}

		if len(data.AlbumArtist) == 0 {
			// Album-level ALBUMARTIST not set, but tracks have it - parse from track value
			data.AlbumArtist = e.parseArtistField(trackAlbumArtistStr)
		} else {
			// Compare formatted strings
			albumArtistStr := domain.FormatArtists(data.AlbumArtist)
			if albumArtistStr != trackAlbumArtistStr {
				// Album-level and track-level ALBUMARTIST differ
				result.Errors = append(result.Errors, ExtractionError{
					Field:    "album_artist",
					Message:  fmt.Sprintf("album-level ALBUMARTIST '%s' differs from track-level '%s'", albumArtistStr, trackAlbumArtistStr),
					Required: false,
				})
			}
		}
	}

	// If album artist is already set (from tags), refine roles using universal performers from tracks
	// This ensures we have accurate roles based on actual track performers
	if len(data.AlbumArtist) > 0 && len(data.Tracks) > 0 {
		universalArtists := domain.DetermineAlbumArtistFromAlbum(data)
		if len(universalArtists) > 0 {
			// Use universal performers (they have correct roles from tracks)
			// Compare names to ensure they match what we parsed from tags
			parsedNames := make(map[string]bool)
			for _, artist := range data.AlbumArtist {
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
				data.AlbumArtist = universalArtists
			}

			// Ensure AlbumArtist performers are present on each track (unless Various Artists)
			if !strings.EqualFold(strings.TrimSpace(domain.FormatArtists(data.AlbumArtist)), "Various Artists") {
				ensureArtistsOnTracks(data.Tracks, data.AlbumArtist)
			}
		}
	}

	// If album artist is empty but we have performers in all tracks, synthesize it
	if len(data.AlbumArtist) == 0 && len(data.Tracks) > 0 {
		universalArtists := domain.DetermineAlbumArtistFromAlbum(data)
		if len(universalArtists) > 0 {
			data.AlbumArtist = universalArtists
			// Ensure AlbumArtist performers are present on each track (unless Various Artists)
			if !strings.EqualFold(strings.TrimSpace(domain.FormatArtists(data.AlbumArtist)), "Various Artists") {
				ensureArtistsOnTracks(data.Tracks, data.AlbumArtist)
			}
		}
	}

	// Try to extract folder name metadata if album title missing
	if data.Title == MissingTitle {
		if folderName, title, year := e.parseDirectoryName(dirPath); folderName != "" && title != "" {
			data.FolderName = folderName
			data.Title = title
			if year > 0 && data.OriginalYear == MissingYear {
				data.OriginalYear = year
			}
			parsingNotes["title_source"] = "directory_name"
			result.Warnings = append(result.Warnings, "album title extracted from directory name")
		}
	}

	// Check for missing required fields
	if data.Title == MissingTitle {
		result.Warnings = append(result.Warnings, "title not found in tags or directory name")
	}
	if data.OriginalYear == MissingYear {
		result.Warnings = append(result.Warnings, "year not found in tags or directory name")
	}

	result.Notes = append(result.Notes, fmt.Sprintf("tracks extracted: %d", len(data.Tracks)))

	// Convert Album to Torrent before returning
	result.Torrent = data.ToTorrent(filepath.Base(dirPath))

	return result, nil
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
func (e *LocalExtractor) extractAlbumMetadata(filePath string) (albumMetadata, string) {
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
	vorbisTags := e.readVorbisCommentTags(filePath)
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
		meta.AlbumArtist = e.parseArtistField(albumArtistStr)
	}

	// Extract edition info - prioritize direct tags, fall back to COMMENT parsing
	edition := e.extractEditionFromTags(vorbisTags)
	if edition == nil {
		// Fall back to COMMENT field parsing
		if comment := metadata.Comment(); comment != "" {
			edition = e.extractEditionFromComment(comment)
		}
	}
	if edition != nil {
		meta.Edition = edition
	}

	return meta, ""
}

// readVorbisCommentTags reads all Vorbis comment tags from a FLAC file.
// Returns a map of tag names (uppercase) to values.
func (e *LocalExtractor) readVorbisCommentTags(filePath string) map[string]string {
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
func (e *LocalExtractor) extractEditionFromTags(tags map[string]string) *domain.Edition {
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
func (e *LocalExtractor) extractEditionFromComment(comment string) *domain.Edition {
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
func (e *LocalExtractor) extractTrackMetadataWithAlbumArtist(filePath string, baseDir string) (*domain.Track, string, error) {
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
		if num := e.extractTrackNumberFromFilename(filePath); num > 0 {
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
		track.Disc = e.extractDiscFromPath(filePath)
	}

	// Extract title
	if title := metadata.Title(); title != "" {
		track.Title = title
	} else {
		// Use filename without extension as fallback
		track.Title = e.extractTitleFromFilename(filePath)
	}

	// Extract composer (required field)
	if composer := metadata.Composer(); composer != "" {
		track.Artists = append(track.Artists, domain.Artist{Name: composer, Role: domain.RoleComposer})
	} else {
		return track, "", fmt.Errorf("no composer found in tags")
	}

	// Extract artists
	if artist := metadata.Artist(); artist != "" {
		track.Artists = append(track.Artists, e.parseArtistField(artist)...)
	} else if albumArtist := metadata.AlbumArtist(); albumArtist != "" {
		// Fallback to album artist if artist tag missing
		track.Artists = append(track.Artists, e.parseArtistField(albumArtist)...)
	}

	// Extract ALBUMARTIST value for verification (but don't store in track)
	albumArtistValue := metadata.AlbumArtist()

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
func (e *LocalExtractor) extractTrackNumberFromFilename(filePath string) int {
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
func (e *LocalExtractor) extractDiscFromPath(filePath string) int {
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
func (e *LocalExtractor) extractTitleFromFilename(filePath string) string {
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

// parseArtistField parses the artist tag field into individual artists.
// Handles formats like "Soloist; Orchestra; Conductor" or "Soloist, Orchestra, Conductor"
// Returns immutable slice.
func (e *LocalExtractor) parseArtistField(artistField string) []domain.Artist {
	artists := make([]domain.Artist, 0)

	// Try semicolon separator first (more reliable)
	var names []string
	if strings.Contains(artistField, ";") {
		names = strings.Split(artistField, ";")
	} else if strings.Contains(artistField, ",") {
		names = strings.Split(artistField, ",")
	} else {
		// Single artist
		names = []string{artistField}
	}

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// Do not infer roles from names; preserve original order and mark as Unknown
		artists = append(artists, domain.Artist{
			Name: name,
			Role: domain.RoleUnknown,
		})
	}

	return artists
}

// inferRoleFromName attempts to infer artist role from their name.
func (e *LocalExtractor) inferRoleFromName(name string) domain.Role {
	nameLower := strings.ToLower(name)

	// Check for explicit role indicators
	if strings.Contains(nameLower, "conductor") || strings.Contains(nameLower, "dir.") {
		return domain.RoleConductor
	}

	// Check for ensemble indicators
	ensembleKeywords := []string{
		"orchestra", "philharmonic", "symphony", "ensemble",
		"choir", "chorus", "kammerchor", "kammer", // German
		"quartet", "trio", "quintet", "sextet",
		"chamber", "band", "consort", "players",
	}
	for _, keyword := range ensembleKeywords {
		if strings.Contains(nameLower, keyword) {
			return domain.RoleEnsemble
		}
	}

	// Default to soloist for individual names
	return domain.RoleSoloist
}

// parseDirectoryName attempts to extract album title and year from directory name.
// Handles formats like "Beethoven - Symphony No. 5 [1963]" or "Bach - Goldberg Variations (1741)"
func (e *LocalExtractor) parseDirectoryName(dirPath string) (folderName string, title string, year int) {
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
