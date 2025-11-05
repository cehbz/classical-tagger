package domain

// Track represents a single track/movement.
// Track embeds File, so it IS a File and can be stored in Files []*File.
// All fields are exported and mutable.
type Track struct {
	File // Embedded - Track IS a File

	// Track-specific metadata
	Disc    int      `json:"disc"`
	Track   int      `json:"track"`
	Title   string   `json:"title"`
	Artists []Artist `json:"artists"`
}

// Composers returns all the composer artists.
func (t *Track) Composers() []*Artist {
	var composers []*Artist
	for _, artist := range t.Artists {
		if artist.Role == RoleComposer {
			composers = append(composers, &artist)
		}
	}
	return composers
}
