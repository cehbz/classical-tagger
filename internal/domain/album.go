package domain

import (
	"fmt"
	"strings"
)

// Album is the aggregate root representing a classical music release.
type Album struct {
	title        string
	originalYear int
	edition      *Edition
	tracks       []*Track
}

// NewAlbum creates a new Album with required fields.
// Returns an error if title is empty or year is invalid.
func NewAlbum(title string, originalYear int) (*Album, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("album title cannot be empty")
	}
	// Year is optional (2.3.16.4). Allow 0 to represent unknown; negative is invalid.
	if originalYear < 0 {
		return nil, fmt.Errorf("album year must be >= 0, got %d", originalYear)
	}

	return &Album{
		title:        title,
		originalYear: originalYear,
		tracks:       make([]*Track, 0),
	}, nil
}

// WithEdition returns a new Album with the edition set.
func (a *Album) WithEdition(edition Edition) *Album {
	a.edition = &edition
	return a
}

// AddTrack adds a track to the album.
func (a *Album) AddTrack(track *Track) error {
	if track == nil {
		return fmt.Errorf("cannot add nil track")
	}
	a.tracks = append(a.tracks, track)
	return nil
}

// Title returns the album title.
func (a *Album) Title() string {
	return a.title
}

// OriginalYear returns the year of original recording/release.
func (a *Album) OriginalYear() int {
	return a.originalYear
}

// Edition returns the edition information, or nil if not set.
func (a *Album) Edition() *Edition {
	return a.edition
}

// Tracks returns a copy of the tracks slice.
func (a *Album) Tracks() []*Track {
	result := make([]*Track, len(a.tracks))
	copy(result, a.tracks)
	return result
}

// Validate checks the entire album for compliance with rules.
// Returns all validation issues from the album and all its tracks.
func (a *Album) Validate() []ValidationIssue {
	var issues []ValidationIssue

	// Album must have at least one track (2.3.16.4 implies music content)
	if len(a.tracks) == 0 {
		issues = append(issues, NewIssue(
			LevelError,
			0, // album-level
			"2.3.16.4",
			"Album must have at least one track",
		))
	}

	// Edition is optional but strongly recommended (Classical Guide preamble)
	if a.edition == nil {
		issues = append(issues, NewIssue(
			LevelWarning,
			0, // album-level
			"Classical Guide: Step 3",
			"Edition information (label, catalog number) is strongly recommended",
		))
	} else {
		// Validate edition if present
		editionIssues := a.edition.Validate()
		issues = append(issues, editionIssues...)
	}

	// Year tag is optional but strongly encouraged (2.3.16.4).
	// Warn if unknown, basic sanity otherwise can be checked elsewhere if needed.
	if a.originalYear == 0 {
		issues = append(issues, NewIssue(
			LevelWarning,
			0,
			"2.3.16.4",
			"Year is optional but strongly encouraged; consider adding original release year",
		))
	}

	// Validate all tracks
	for _, track := range a.tracks {
		trackIssues := track.Validate()
		issues = append(issues, trackIssues...)
	}

	return issues
}
