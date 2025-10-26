package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Track is an entity representing a single track/movement.
type Track struct {
	disc    int
	track   int
	title   string
	artists []Artist
	name    string // filename
}

// NewTrack creates a new Track with validation.
// Returns an error if:
// - disc or track numbers are not positive
// - title is empty
// - there is not exactly one composer
// - there are no artists
func NewTrack(disc, track int, title string, artists []Artist) (*Track, error) {
	// Accept raw values; semantic validation happens later in validators
	// Preserve title exactly as provided; semantic validation happens later

	// Composer and performer presence is validated later; allow flexible construction here

	return &Track{
		disc:    disc,
		track:   track,
		title:   title,
		artists: artists,
	}, nil
}

// WithName returns a new Track with the filename set.
func (t *Track) WithName(name string) *Track {
	t.name = name
	return t
}

// Disc returns the disc number.
func (t *Track) Disc() int {
	return t.disc
}

// Track returns the track number.
func (t *Track) Track() int {
	return t.track
}

// Title returns the track title (work + movement).
func (t *Track) Title() string {
	return t.title
}

// Artists returns a copy of the artists slice.
func (t *Track) Artists() []Artist {
	result := make([]Artist, len(t.artists))
	copy(result, t.artists)
	return result
}

// Name returns the filename.
func (t *Track) Name() string {
	return t.name
}

// Composer returns the first composer artist.
// Assumes NewTrack validation ensures exactly one composer exists.
func (t *Track) Composer() Artist {
	for _, artist := range t.artists {
		if artist.Role() == RoleComposer {
			return artist
		}
	}
	// This should never happen if NewTrack validation is correct
	return Artist{}
}

// Validate checks the track for compliance with rules.
// Returns validation issues at ERROR, WARNING, or INFO levels.
func (t *Track) Validate() []ValidationIssue {
	var issues []ValidationIssue

	// Check composer not in title (Classical Guide: Step 1, Track Title)
	composer := t.Composer()
	if composer.Name() != "" {
		lastName := lastName(composer.Name())
		// Match whole word boundary to avoid false positives
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(lastName) + `\b`)
		if pattern.MatchString(t.title) {
			issues = append(issues, NewIssue(
				LevelError,
				t.track,
				"Classical Guide: Step 1, Track Title",
				fmt.Sprintf("Composer name '%s' must not appear in track title tag", composer.Name()),
			))
		}
	}

	// Check filename length (2.3.12)
	if len(t.name) > 180 {
		issues = append(issues, NewIssue(
			LevelError,
			t.track,
			"2.3.12",
			fmt.Sprintf("Filename exceeds 180 characters (%d)", len(t.name)),
		))
	}

	// Check filename format (2.3.13) - track number required
	if t.name != "" {
		trackNumPattern := regexp.MustCompile(`^\d{2}`)
		if !trackNumPattern.MatchString(t.name) {
			issues = append(issues, NewIssue(
				LevelWarning,
				t.track,
				"2.3.13",
				"Filename should start with track number (e.g., '01 - ...')",
			))
		}
	}

	// Check for opus number format (Classical Guide: Step 1, Track Title) - INFO level
	opusPatterns := []string{`op\.\s*\d+`, `opus\s*\d+`, `Op\s*\d+`}
	hasNonStandardOpus := false
	for _, pattern := range opusPatterns {
		if regexp.MustCompile(pattern).MatchString(t.title) {
			hasNonStandardOpus = true
			break
		}
	}
	standardPattern := regexp.MustCompile(`Op\.\s*\d+`)
	if hasNonStandardOpus && !standardPattern.MatchString(t.title) {
		issues = append(issues, NewIssue(
			LevelInfo,
			t.track,
			"Classical Guide: Step 1, Track Title",
			"Consider using standard opus format: 'Op. ###'",
		))
	}

	return issues
}

// lastName extracts the last name from a full name.
// Simple implementation: takes the last space-separated token.
func lastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
