package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Track represents a single track/movement.
// All fields are exported and mutable.
type Track struct {
	Disc    int      `json:"disc"`
	Track   int      `json:"track"`
	Title   string   `json:"title"`
	Name    string   `json:"name,omitempty"` // filename
	Artists []Artist `json:"artists"`
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

	// Rule classical.track_title: Composer surname must not appear in track title
	for _, composer := range t.Composers() {
		if composer.Name == "" || t.Title == "" {
			continue
		}

		// Extract composer last name(s)
		lastNames := extractLastNames(composer.Name)
		titleLower := strings.ToLower(t.Title)

		for _, lastName := range lastNames {
			lastNameLower := strings.ToLower(lastName)

			// Use word boundary pattern to avoid false positives
			// Match: "Bach: Symphony" or "Symphony (Bach)"
			// Don't match: "Bacharach" (different word)
			pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(lastNameLower) + `\b`)

			if pattern.MatchString(titleLower) {
				// Check for exceptions: "Variations on a Theme by Brahms" is acceptable
				if isComposerPartOfWorkTitle(t.Title, lastName) {
					continue
				}

				issues = append(issues, ValidationIssue{
					Level:   LevelError,
					Track:   t.Track,
					Rule:    "classical.track_title",
					Message: fmt.Sprintf("Composer surname '%s' found in track title (composer should only be in COMPOSER tag)", lastName),
				})
			}
		}
	}

	// Rule 2.3.12: Path length must not exceed 180 characters
	if t.Name != "" && len(t.Name) > 180 {
		issues = append(issues, ValidationIssue{
			Level:   LevelError,
			Track:   t.Track,
			Rule:    "2.3.12",
			Message: fmt.Sprintf("Filename exceeds 180 characters (%d)", len(t.Name)),
		})
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

// extractLastNames extracts last name(s) from a composer name
// "Ludwig van Beethoven" -> ["Beethoven"]
// "J.S. Bach" -> ["Bach"]
// "Beethoven, Ludwig van" -> ["Beethoven"]
func extractLastNames(composerName string) []string {
	// Handle reversed format: "Beethoven, Ludwig van"
	if strings.Contains(composerName, ",") {
		parts := strings.Split(composerName, ",")
		return []string{strings.TrimSpace(parts[0])}
	}

	// Handle normal format: "Ludwig van Beethoven"
	parts := strings.Fields(composerName)
	if len(parts) == 0 {
		return []string{}
	}

	// Handle compound last names with lowercase particles: "van", "von", "de", "da"
	// "Ludwig van Beethoven" -> look back for "van" and include it
	var lastName string
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if lastName == "" {
			lastName = part
		} else {
			// Check if this is a lowercase particle
			if part == strings.ToLower(part) && len(part) < 4 {
				lastName = part + " " + lastName
			} else {
				// Found a capitalized name, stop here
				break
			}
		}
	}

	return []string{lastName}
}

// isComposerPartOfWorkTitle checks if composer name is part of the actual work title
// "Variations on a Theme by Brahms" -> true (Brahms is part of work title)
// "Symphony No. 5 - Brahms" -> false (Brahms is wrongly appended)
func isComposerPartOfWorkTitle(title, composerLastName string) bool {
	// Common patterns where composer is legitimately part of the work title
	patterns := []string{
		"on a theme by " + strings.ToLower(composerLastName),
		"variations on " + strings.ToLower(composerLastName),
		"after " + strings.ToLower(composerLastName),
		"hommage to " + strings.ToLower(composerLastName),
		"hommage à " + strings.ToLower(composerLastName),
		"in memory of " + strings.ToLower(composerLastName),
	}

	titleLower := strings.ToLower(title)
	for _, pattern := range patterns {
		if strings.Contains(titleLower, pattern) {
			return true
		}
	}

	return false
}

// isValidFilename checks if a filename matches the required format.
// Format: NN Title.ext or NN - Title.ext (where NN is track number)
func isValidFilename(name string) bool {
	// Pattern: starts with 1-3 digits, optional separator, then content
	pattern := regexp.MustCompile(`^\d{1,3}[\s\-._]?.*\.\w+$`)
	return pattern.MatchString(name)
}
