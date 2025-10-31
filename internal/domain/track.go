package domain

// Track represents a single track/movement.
// All fields are exported and mutable.
type Track struct {
	Disc    int      `json:"disc"`
	Track   int      `json:"track"`
	Title   string   `json:"title"`
	Name    string   `json:"name,omitempty"` // filename
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