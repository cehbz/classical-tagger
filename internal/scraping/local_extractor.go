package scraping

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

// LocalExtractor extracts metadata from existing FLAC files in a directory.
// This is useful for converting already-tagged albums to our JSON format.
type LocalExtractor struct {
	// Extractor is stateless
}

// NewLocalExtractor creates a new local extractor.
func NewLocalExtractor() *LocalExtractor {
	return &LocalExtractor{}
}

// ExtractFromDirectory reads all FLAC files in a directory and extracts metadata.
// It attempts to build a complete AlbumData structure from the tags and filenames.
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
	data := &AlbumData{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]TrackData, 0, len(files)),
	}

	result := NewExtractionResult(data)
	parsingNotes := make(map[string]interface{})
	parsingNotes["source"] = "local_directory"
	parsingNotes["directory"] = dirPath
	parsingNotes["file_count"] = len(files)

	// Extract album-level metadata from first file
	if len(files) > 0 {
		if err := e.extractAlbumMetadata(files[0], data, result); err != nil {
			result = result.WithWarning(fmt.Sprintf("album metadata extraction: %v", err))
		}
	}

	// Extract track metadata from each file
	for _, filePath := range files {
		track, err := e.extractTrackMetadata(filePath)
		if err != nil {
			result = result.WithWarning(fmt.Sprintf("file %s: %v", filepath.Base(filePath), err))
			continue
		}

		data.Tracks = append(data.Tracks, track)
	}

	// Validate we got tracks
	if len(data.Tracks) == 0 {
		result = result.WithError(NewExtractionError("tracks", "no tracks extracted", true))
	}

	// Try to extract folder name metadata if album title missing
	if data.Title == MissingTitle {
		if title, year := e.parseDirectoryName(dirPath); title != "" {
			data.Title = title
			if year > 0 && data.OriginalYear == MissingYear {
				data.OriginalYear = year
			}
			parsingNotes["title_source"] = "directory_name"
		}
	}

	// Check for missing required fields
	if data.Title == MissingTitle {
		result = result.WithError(NewExtractionError("title", "not found in tags or directory name", true))
	}
	if data.OriginalYear == MissingYear {
		result = result.WithError(NewExtractionError("year", "not found in tags or directory name", true))
	}

	parsingNotes["tracks_extracted"] = len(data.Tracks)
	result = result.WithParsingNotes(parsingNotes)

	return result, nil
}

// extractAlbumMetadata extracts album-level metadata from a FLAC file's tags.
func (e *LocalExtractor) extractAlbumMetadata(filePath string, data *AlbumData, result *ExtractionResult) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		return fmt.Errorf("failed to read tags: %w", err)
	}

	// Extract album title
	if album := metadata.Album(); album != "" {
		data.Title = album
	}

	// Extract year
	if year := metadata.Year(); year > 0 {
		data.OriginalYear = year
	}

	// Extract album artist (may be useful for edition info)
	if albumArtist := metadata.AlbumArtist(); albumArtist != "" {
		// Store in parsing notes for now
		// Could be used to infer ensemble/conductor
	}

	return nil
}

// extractTrackMetadata extracts track-level metadata from a FLAC file.
func (e *LocalExtractor) extractTrackMetadata(filePath string) (TrackData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return TrackData{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	metadata, err := tag.ReadFrom(f)
	if err != nil {
		return TrackData{}, fmt.Errorf("failed to read tags: %w", err)
	}

	track := TrackData{
		Disc:     1, // Default
		Track:    0,
		Title:    "",
		Composer: "",
		Artists:  make([]ArtistData, 0),
	}

	// Extract track number
	trackNum, _ := metadata.Track()
	if trackNum > 0 {
		track.Track = trackNum
	} else {
		// Try to extract from filename
		if num := e.extractTrackNumberFromFilename(filePath); num > 0 {
			track.Track = num
		}
	}

	// Extract disc number
	discNum, _ := metadata.Disc()
	if discNum > 0 {
		track.Disc = discNum
	} else {
		// Try to extract from path (CD1, CD2, etc.)
		if disc := e.extractDiscFromPath(filePath); disc > 0 {
			track.Disc = disc
		}
	}

	// Extract title
	if title := metadata.Title(); title != "" {
		track.Title = title
	} else {
		// Use filename without extension as fallback
		track.Title = e.extractTitleFromFilename(filePath)
	}

	// Extract composer
	if composer := metadata.Composer(); composer != "" {
		track.Composer = composer
	}

	// Extract artist(s)
	if artist := metadata.Artist(); artist != "" {
		// Parse artist field - may contain multiple artists
		artists := e.parseArtistField(artist)
		track.Artists = artists
	}

	// If no track number, this is an error
	if track.Track == 0 {
		return track, fmt.Errorf("no track number found")
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
		pattern := regexp.MustCompile(`(?i)(?:CD|Disc)\s*(\d+)`)
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
func (e *LocalExtractor) parseArtistField(artistField string) []ArtistData {
	artists := make([]ArtistData, 0)

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

		artists = append(artists, ArtistData{
			Name: name,
			Role: role,
		})
	}

	return artists
}

// inferRoleFromName attempts to infer artist role from their name.
func (e *LocalExtractor) inferRoleFromName(name string) string {
	nameLower := strings.ToLower(name)

	// Check for explicit role indicators
	if strings.Contains(nameLower, "conductor") || strings.Contains(nameLower, "dir.") {
		return "conductor"
	}
	if strings.Contains(nameLower, "orchestra") || strings.Contains(nameLower, "philharmonic") {
		return "ensemble"
	}
	if strings.Contains(nameLower, "ensemble") || strings.Contains(nameLower, "quartet") {
		return "ensemble"
	}
	if strings.Contains(nameLower, "choir") || strings.Contains(nameLower, "chorus") {
		return "ensemble"
	}

	// Default to soloist for individual names
	return "soloist"
}

// parseDirectoryName attempts to extract album title and year from directory name.
// Supports formats like:
// - "Composer - Title (Year) - Format"
// - "Composer - Title (Year)"
// - "Title (Year)"
func (e *LocalExtractor) parseDirectoryName(dirPath string) (title string, year int) {
	dirName := filepath.Base(dirPath)

	// Pattern: "... (YYYY) ..."
	yearPattern := regexp.MustCompile(`\((\d{4})\)`)
	matches := yearPattern.FindStringSubmatch(dirName)
	if len(matches) > 1 {
		y, err := strconv.Atoi(matches[1])
		if err == nil && y >= 1900 && y <= 2030 {
			year = y
		}

		// Remove year from title
		dirName = yearPattern.ReplaceAllString(dirName, "")
	}

	// Remove format suffix like "- FLAC", "- 24-96", etc.
	formatPattern := regexp.MustCompile(`\s*-\s*(FLAC|MP3|WAV|ALAC|DSD|\d{2}-\d{2,3}).*$`)
	dirName = formatPattern.ReplaceAllString(dirName, "")

	// Clean up
	title = strings.TrimSpace(dirName)
	title = strings.Trim(title, "-")
	title = strings.TrimSpace(title)

	return title, year
}