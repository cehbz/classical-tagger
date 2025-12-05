package domain

import "strings"

// Artist represents a person involved in a recording.
// All fields are exported and mutable.
type Artist struct {
	Name string `json:"name"`
	Role Role   `json:"role"`
}

// String returns a string representation of the artist (Name - Role).
func (a Artist) String() string {
	return a.Name + " (" + a.Role.String() + ")"
}

// ParseArtist creates an Artist from name and role string.
func ParseArtist(name, roleStr string) (Artist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Artist{}, ErrEmptyArtistName
	}

	role, err := ParseRole(roleStr)
	if err != nil {
		return Artist{}, err
	}

	return Artist{
		Name: name,
		Role: role,
	}, nil
}

// ParseArtistField parses a comma or semicolon-separated artist tag field into individual artists.
// Handles formats like "Soloist; Orchestra; Conductor" or "Soloist, Orchestra, Conductor".
// Returns a slice of artists with RoleUnknown (roles should be inferred from context).
// This is used for parsing FLAC tags where multiple artists may be stored in a single tag.
func ParseArtistField(artistField string) []Artist {
	artists := make([]Artist, 0)

	// Try semicolon separator first (more reliable)
	names := strings.Split(artistField, ";")
	if len(names) == 1 {
		names = strings.Split(artistField, ",")
	}

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// Do not infer roles from names; preserve original order and mark as Unknown
		// Roles should be inferred from context (e.g., ALBUMARTIST vs ARTIST vs COMPOSER tags)
		artists = append(artists, Artist{
			Name: name,
			Role: RoleUnknown,
		})
	}

	return artists
}
