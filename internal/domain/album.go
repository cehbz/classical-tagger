package domain

// Album represents a classical music release.
type Album struct {
	FolderName   string   `json:"folder_name"`
	Title        string   `json:"title"`
	OriginalYear int      `json:"original_year"`
	Edition      *Edition `json:"edition,omitempty"`
	AlbumArtist  []Artist `json:"album_artist,omitempty"`
	Tracks       []*Track `json:"tracks"`
}

// IsMultiDisc returns true if the album contains tracks from multiple discs.
// An album is considered multi-disc if any track has Disc > 1 or if there are multiple distinct disc numbers.
func (a Album) IsMultiDisc() bool {
	if len(a.Tracks) == 0 {
		return false
	}

	maxDisc := 1
	discSet := make(map[int]bool)
	for _, track := range a.Tracks {
		if track.Disc > maxDisc {
			maxDisc = track.Disc
		}
		discSet[track.Disc] = true
	}

	// Multi-disc if max disc > 1 OR if there are multiple distinct disc numbers
	return maxDisc > 1 || len(discSet) > 1
}

// DetermineAlbumArtistFromAlbum finds performers that appear in all tracks of an Album.
// This is a compatibility function that will be deprecated in favor of DetermineAlbumArtist(torrent *Torrent).
// Returns the list of universal artists.
// Per classical music guide: "When the performer(s) do not remain the same throughout
// all tracks, this tag is used to credit the one who does appear in all tracks."
func (album Album) AlbumArtists() []Artist {
	tracks := album.Tracks
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

// ToTorrent converts an Album to a Torrent.
// The rootPath parameter should be the relative path to the torrent directory.
// For each track, the Path field will be set from the track's Name field if present.
func (a *Album) ToTorrent(rootPath string) *Torrent {
	if a == nil {
		return nil
	}

	fs := make([]FileLike, len(a.Tracks))
	for i, tr := range a.Tracks {
		fs[i] = tr
	}

	return &Torrent{
		RootPath:     rootPath,
		Title:        a.Title,
		OriginalYear: a.OriginalYear,
		Edition:      a.Edition,
		AlbumArtist:  a.AlbumArtist,
		Files:        fs,
		SiteMetadata: nil, // Not available from Album
	}
}