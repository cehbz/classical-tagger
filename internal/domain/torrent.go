package domain

import (
	"encoding/json"
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

// DetermineAlbumArtist finds performers that appear in all tracks.
// Returns the list of universal artists.
// Per classical music guide: "When the performer(s) do not remain the same throughout
// all tracks, this tag is used to credit the one who does appear in all tracks."
func DetermineAlbumArtist(torrent *Torrent) []Artist {
	tracks := torrent.Tracks()
	if len(tracks) == 0 {
		return nil
	}

	// Build set of all non-composer artists from first track
	firstTrack := tracks[0]
	var candidates []Artist
	for _, artist := range firstTrack.Artists {
		if artist.Role != RoleComposer {
			candidates = append(candidates, artist)
		}
	}

	// Filter to only those appearing in ALL tracks
	var universal []Artist
	for _, candidate := range candidates {
		appearsInAll := true
		for _, track := range tracks[1:] {
			found := false
			for _, artist := range track.Artists {
				if artist.Name == candidate.Name && artist.Role == candidate.Role {
					found = true
					break
				}
			}
			if !found {
				appearsInAll = false
				break
			}
		}
		if appearsInAll {
			universal = append(universal, candidate)
		}
	}

	return universal
}
