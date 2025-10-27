package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Track represents a single track/movement.
// All fields are exported and mutable.
type Track struct {
	Disc    int       `json:"disc"`
	Track   int       `json:"track"`
	Title   string    `json:"title"`
	Name    string    `json:"name,omitempty"` // filename
	Artists []Artist  `json:"artists"`
}

// Composer returns all the composer artists.
func (t *Track) Composers() []*Artist {
	var composers []*Artist
	for _, artist := range t.Artists {
		if artist.Role == RoleComposer {
			composers = append(composers, &artist)
		}
	}
	// Return empty artist if no composer found
	return composers
}

// Validate checks the track for compliance with rules.
// Returns validation issues at ERROR, WARNING, or INFO levels.
func (t *Track) Validate() []ValidationIssue {
	var issues []ValidationIssue

	// Rule 2.3.8: Composer name must not appear in track title
	for _, composer := range t.Composers() {
		if composer.Name != "" && strings.Contains(t.Title, composer.Name) {
			issues = append(issues, ValidationIssue{
				Level:   LevelError,
				Track:   t.Track,
				Rule:    "2.3.8",
				Message: fmt.Sprintf("Composer name '%s' must not appear in track title tag", composer.Name),
			})
		}
	}

	// Rule 2.3.13: Filename format validation
	if t.Name != "" {
		if !isValidFilename(t.Name) {
			issues = append(issues, ValidationIssue{
				Level:   LevelError,
				Track:   t.Track,
				Rule:    "2.3.13",
				Message: fmt.Sprintf("Filename '%s' does not match required format", t.Name),
			})
		}
	}

	// Check for arranger in title (optional parsing)
	if strings.Contains(strings.ToLower(t.Title), "arr.") ||
		strings.Contains(strings.ToLower(t.Title), "arranged by") {
		issues = append(issues, ValidationIssue{
			Level:   LevelInfo,
			Track:   t.Track,
			Rule:    "2.3.7",
			Message: "Track title contains arranger information - consider using ARRANGER tag",
		})
	}

	return issues
}

// isValidFilename checks if a filename matches the required format.
// Format: NN Title.ext or NN - Title.ext (where NN is track number)
func isValidFilename(name string) bool {
	// Pattern: starts with 1-3 digits, optional separator, then content
	pattern := regexp.MustCompile(`^\d{1,3}[\s\-._]?.*\.\w+$`)
	return pattern.MatchString(name)
}
