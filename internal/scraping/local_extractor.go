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

	// Start with empty result
	result := &ExtractionResult{
		Album:  data,
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

		if warning != "" {
			result.Warnings = append(result.Warnings, warning)
		}
	}

	// Extract track metadata from each file
	for _, filePath := range files {
		track, err := e.extractTrackMetadata(filePath, dirPath)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("file %s: %v", filepath.Base(filePath), err))
			continue
		}

		data.Tracks = append(data.Tracks, track)
	}

	// Validate we got tracks
	if len(data.Tracks) == 0 {
		result.Warnings = append(result.Warnings, "no tracks extracted")
		return result, nil
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

	return result, nil
}

// albumMetadata is a temporary structure for album-level data
type albumMetadata struct {
	Title        string
	OriginalYear int
	Edition      *domain.Edition
}

// extractAlbumMetadata extracts album-level metadata from a FLAC file's tags.
// Returns the metadata and an optional warning string.
func (e *LocalExtractor) extractAlbumMetadata(filePath string) (albumMetadata, string) {
	meta := albumMetadata{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
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

	// Extract year
	if year := metadata.Year(); year > 0 {
		meta.OriginalYear = year
	}

	// Extract edition info if present in comments
	if comment := metadata.Comment(); comment != "" {
		edition := e.extractEditionFromComment(comment)
		if edition != nil {
			meta.Edition = edition
		}
	}

	return meta, ""
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

// extractTrackMetadata extracts track-level metadata from a FLAC file.
func (e *LocalExtractor) extractTrackMetadata(filePath string, baseDir string) (*domain.Track, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags: %w", err)
	}

	track := &domain.Track{
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
			return track, fmt.Errorf("no track number found in tags or filename")
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
		return track, fmt.Errorf("no composer found in tags")
	}

	// Extract artists
	if artist := metadata.Artist(); artist != "" {
		track.Artists = e.parseArtistField(artist)
	} else if albumArtist := metadata.AlbumArtist(); albumArtist != "" {
		track.Artists = e.parseArtistField(albumArtist)
	}

	// Set relative filename (add before the final return)
	relPath, err := filepath.Rel(baseDir, filePath)
	if err == nil {
		// Convert to forward slashes for consistency
		track.Name = filepath.ToSlash(relPath)
	} else {
		track.Name = filepath.Base(filePath)
	}

	return track, nil
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

		// Try to infer role from name
		role := e.inferRoleFromName(name)

		artists = append(artists, domain.Artist{
			Name: name,
			Role: domain.Role(role),
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
