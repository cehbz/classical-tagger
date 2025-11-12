package domain

import (
	"encoding/json"
	"fmt"
	"os"
)

// Torrent represents a torrent directory with associated metadata and files.
// It is the aggregate root of the domain model.
type Torrent struct {
	// Location on disk
	RootPath string `json:"root_path"` // Relative path to torrent directory

	// Album-level metadata
	Title        string   `json:"title"`
	OriginalYear int      `json:"original_year"`
	Edition      *Edition `json:"edition,omitempty"`
	AlbumArtist  []Artist `json:"album_artist,omitempty"`

	// All files in the torrent (mix of File and Track)
	Files []FileLike `json:"files"`

	// Site-specific metadata (optional, for upload)
	SiteMetadata *SiteMetadata `json:"site_metadata,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for Torrent.
// This is needed because Files contains FileLike interface values which need to be
// marshaled as their concrete types (File or Track).
func (t *Torrent) MarshalJSON() ([]byte, error) {
	type torrentJSON struct {
		RootPath     string        `json:"root_path"`
		Title        string        `json:"title"`
		OriginalYear int           `json:"original_year"`
		Edition      *Edition      `json:"edition,omitempty"`
		AlbumArtist  []Artist      `json:"album_artist,omitempty"`
		Files        any           `json:"files"`
		SiteMetadata *SiteMetadata `json:"site_metadata,omitempty"`
	}

	// Marshal Files array by converting each FileLike to its concrete type
	filesData := make([]any, 0, len(t.Files))
	for _, fileLike := range t.Files {
		// Check concrete type and add the concrete value
		// FileLike interface is only satisfied by pointer types (*File, *Track)
		switch v := fileLike.(type) {
		case *Track:
			filesData = append(filesData, v)
		case *File:
			filesData = append(filesData, v)
		default:
			// Fallback: try to marshal as-is
			filesData = append(filesData, fileLike)
		}
	}

	tj := torrentJSON{
		RootPath:     t.RootPath,
		Title:        t.Title,
		OriginalYear: t.OriginalYear,
		Edition:      t.Edition,
		AlbumArtist:  t.AlbumArtist,
		Files:        filesData,
		SiteMetadata: t.SiteMetadata,
	}

	return json.Marshal(tj)
}

// UnmarshalJSON implements custom JSON unmarshaling for Torrent.
// This is needed because Files contains FileLike interface values which need to be
// unmarshaled as either File or Track based on JSON content.
func (t *Torrent) UnmarshalJSON(data []byte) error {
	// Use an intermediate struct with Files as raw JSON
	type torrentJSON struct {
		RootPath     string          `json:"root_path"`
		Title        string          `json:"title"`
		OriginalYear int             `json:"original_year"`
		Edition      *Edition        `json:"edition,omitempty"`
		AlbumArtist  []Artist        `json:"album_artist,omitempty"`
		Files        json.RawMessage `json:"files"`
		SiteMetadata *SiteMetadata   `json:"site_metadata,omitempty"`
	}

	var tmp torrentJSON
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// Copy simple fields
	t.RootPath = tmp.RootPath
	t.Title = tmp.Title
	t.OriginalYear = tmp.OriginalYear
	t.Edition = tmp.Edition
	t.AlbumArtist = tmp.AlbumArtist
	t.SiteMetadata = tmp.SiteMetadata

	// Unmarshal Files array (Files field may be missing or null)
	if len(tmp.Files) > 0 {
		var filesArray []json.RawMessage
		if err := json.Unmarshal(tmp.Files, &filesArray); err != nil {
			return err
		}

		t.Files = make([]FileLike, 0, len(filesArray))
		for _, fileData := range filesArray {
			// Try to unmarshal as Track first (has disc, track, title, artists fields)
			var track Track
			if err := json.Unmarshal(fileData, &track); err == nil {
				// Check if it has Track-specific fields (not just File fields)
				if track.Disc > 0 || track.Track > 0 || track.Title != "" || len(track.Artists) > 0 {
					t.Files = append(t.Files, &track)
					continue
				}
			}

			// Otherwise unmarshal as File
			var file File
			if err := json.Unmarshal(fileData, &file); err != nil {
				return err
			}
			t.Files = append(t.Files, &file)
		}
	} else {
		// Files field is missing or null - initialize empty slice
		t.Files = []FileLike(nil)
	}

	return nil
}

// SaveToFile saves the torrent to a file.
func (t *Torrent) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(t)
}

// IsMultiDisc returns true if the torrent contains tracks from multiple discs.
// A torrent is considered multi-disc if any track has Disc > 1 or if there are multiple distinct disc numbers.
func (t *Torrent) IsMultiDisc() bool {
	tracks := t.Tracks()
	if len(tracks) == 0 {
		return false
	}

	maxDisc := 1
	discSet := make(map[int]bool)
	for _, track := range tracks {
		if track.Disc > maxDisc {
			maxDisc = track.Disc
		}
		discSet[track.Disc] = true
	}

	// Multi-disc if max disc > 1 OR if there are multiple distinct disc numbers
	return maxDisc > 1 || len(discSet) > 1
}

// Tracks returns all files that are tracks (extracts Track instances from Files slice).
// Uses reflection to check if a *File is actually a *Track.
func (t *Torrent) Tracks() []*Track {
	var tracks []*Track
	for _, file := range t.Files {
		if track, ok := file.(*Track); ok {
			tracks = append(tracks, track)
		}
	}
	return tracks
}

// AlbumArtists finds performers that appear in all tracks.
// Returns the list of universal artists.
// Per classical music guide: "When the performer(s) do not remain the same throughout
// all tracks, this tag is used to credit the one who does appear in all tracks."
func (torrent Torrent) AlbumArtists() []Artist {
	artistCounts := make(map[string]int)
	artistOrder := make([]Artist, 0)
	for _, track := range torrent.Tracks() {
		trackArtists := make(map[string]struct{})
		for _, artist := range track.Artists {
			if !artist.Role.IsPerformer() {
				continue
			}
			if _, ok := trackArtists[artist.Name]; ok {
				continue
			}
			trackArtists[artist.Name] = struct{}{}
			if artistCounts[artist.Name] == 0 {
				artistOrder = append(artistOrder, artist)
			}
			artistCounts[artist.Name]++
		}
	}

	var albumArtists []Artist
	for _, artist := range artistOrder {
		if artistCounts[artist.Name] == len(torrent.Tracks()) {
			albumArtists = append(albumArtists, artist)
		}
	}
	return albumArtists
}

// DirectoryName generates a directory name for a torrent following classical music conventions.
// Format: "Composer - Album Title (Performers) - Year [FLAC]"
// Falls back to simpler formats if too long.
// Minimum: "Album Title" (rule 2.3.2)
func (torrent Torrent) DirectoryName() string {
	// Get album title
	albumTitle := SanitizeDirectoryName(torrent.Title)
	if albumTitle == "" {
		albumTitle = "Untitled Album"
	}
	dirName := albumTitle
	dirNameLen := len(dirName)

	formatIndicator := " [FLAC]"
	if dirNameLen+len(formatIndicator) > 180 {
		return dirName
	}
	dirNameLen += len(formatIndicator)
	yearStr := ""
	// Get year
	if torrent.OriginalYear > 0 {
		yearStr = fmt.Sprintf(" - %d", torrent.OriginalYear)
	}
	if dirNameLen+len(yearStr) > 180 {
		return dirName + formatIndicator
	}
	dirNameLen += len(yearStr)

	// Get primary composer(s) - prefer AlbumArtist, fall back to tracks only if AlbumArtist is empty
	// If AlbumArtist is set but has no composers, skip composer prefix (for Discogs releases with only performers)
	composers := torrent.Composers()
	if len(composers) == 0 && len(torrent.AlbumArtist) == 0 {
		// Only fall back to tracks if AlbumArtist is completely empty (local extraction)
		composers = torrent.PrimaryComposers()
	}
	composerStr := ""
	if len(composers) > 0 {
		composerStr = formatComposersForDirectory(composers) + " - "
		if dirNameLen+len(composerStr) > 180 {
			return dirName + yearStr + formatIndicator
		}
		dirName = composerStr + dirName
		dirNameLen += len(composerStr)
	}

	// Get performers (for optional inclusion) - prefer AlbumArtist, fall back to tracks
	performers := torrent.Performers()
	if len(performers) == 0 {
		performers = torrent.PrimaryPerformers()
	}
	performerStr := ""
	if len(performers) > 0 {
		performerStr = " (" + formatPerformersForDirectory(performers) + ")"
	}
	if dirNameLen+len(performerStr) > 180 {
		return dirName + yearStr + formatIndicator
	}
	return dirName + performerStr + yearStr + formatIndicator
}

// Composers extracts composer names from AlbumArtist.
func (t Torrent) Composers() []string {
	var composers []string
	for _, artist := range t.AlbumArtist {
		if artist.Role == RoleComposer && artist.Name != "" {
			composers = append(composers, artist.Name)
		}
	}
	return composers
}

// Performers extracts performer names (non-composers) from AlbumArtist.
func (t Torrent) Performers() []string {
	var performers []string
	for _, artist := range t.AlbumArtist {
		if artist.Role != RoleComposer && artist.Name != "" {
			performers = append(performers, artist.Name)
		}
	}
	return performers
}

// PrimaryComposers extracts the primary composer from tracks.
// Returns the most frequent composer, or empty string if no single composer appears on more than half the tracks.
func (t Torrent) PrimaryComposers() []string {
	composerCounts := make(map[string]int)
	composerOrder := make([]string, 0)

	for _, track := range t.Tracks() {
		for _, artist := range track.Artists {
			if artist.Role == RoleComposer && artist.Name != "" {
				if composerCounts[artist.Name] == 0 {
					composerOrder = append(composerOrder, artist.Name)
				}
				composerCounts[artist.Name]++
			}
		}
	}

	// Find most frequent composers
	var primaryComposers []string
	for _, name := range composerOrder {
		if composerCounts[name] > len(t.Tracks())/2 {
			primaryComposers = append(primaryComposers, name)
		}
	}
	return primaryComposers
}

// PrimaryPerformers extracts primary performers (non-composers) from tracks.
// Returns performers that appear in most tracks.
func (t Torrent) PrimaryPerformers() []string {
	performerCounts := make(map[string]int)
	performerOrder := make([]string, 0)

	for _, track := range t.Tracks() {
		for _, artist := range track.Artists {
			if artist.Role.IsPerformer() && artist.Name != "" {
				if performerCounts[artist.Name] == 0 {
					performerOrder = append(performerOrder, artist.Name)
				}
				performerCounts[artist.Name]++
			}
		}
	}

	// Return performers that appear in at least 50% of tracks
	var result []string
	for _, name := range performerOrder {
		if performerCounts[name] > len(t.Tracks())/2 {
			result = append(result, name)
		}
	}

	return result
}
