package domain

import (
	"fmt"
	"strings"
)

// Artist is a value object representing a performer, composer, conductor, etc.
// It is immutable after creation.
type Artist struct {
	name string
	role Role
}

// NewArtist creates a new Artist with the given name and role.
// Returns an error if the name is empty or whitespace-only.
func NewArtist(name string, role Role) (Artist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Artist{}, fmt.Errorf("artist name cannot be empty")
	}
	
	return Artist{
		name: name,
		role: role,
	}, nil
}

// Name returns the artist's name.
func (a Artist) Name() string {
	return a.name
}

// Role returns the artist's role.
func (a Artist) Role() Role {
	return a.role
}
