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
