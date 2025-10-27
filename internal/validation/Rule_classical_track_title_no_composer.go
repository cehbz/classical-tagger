package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ComposerNotInTitle checks that composer name is NOT in track title (classical.track_title)
// The composer should only be in the COMPOSER tag, not repeated in the title
func (r *Rules) ComposerNotInTitle(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.track_title",
		Name:   "Composer name not in track title",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	for _, track := range actual.Tracks {
		// Find the composer
		var composer *domain.Artist
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleComposer {
				composer = &artist
				break
			}
		}

		if composer == nil {
			// No composer to check - this will be caught by ComposerTagRequired rule
			continue
		}

		title := track.Title
		if title == "" {
			// Empty title - will be caught by RequiredTags rule
			continue
		}

		// Extract composer last name(s)
		composerLastNames := extractLastNames(composer.Name)

		// Check if any last name appears in the title
		// Use word boundaries to avoid false positives
		titleLower := strings.ToLower(title)
		for _, lastName := range composerLastNames {
			lastNameLower := strings.ToLower(lastName)

			// Create a word boundary pattern
			// Match: "Bach: Symphony" or "Symphony (Bach)" or "Bach's"
			// Don't match: "Bacharach" (different word)
			pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(lastNameLower) + `\b`)

			if pattern.MatchString(titleLower) {
				// Check for exceptions: "Variations on a Theme by Brahms" is acceptable
				// These are part of the actual work title, not just composer mention
				if isComposerPartOfWorkTitle(title, lastName) {
					continue
				}

				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelError,
					Track: track.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s: Composer surname '%s' found in track title '%s' (composer should only be in COMPOSER tag)",
						formatTrackNumber(track), lastName, title),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// extractLastNames extracts last name(s) from a composer name
// "Ludwig van Beethoven" -> ["Beethoven"]
// "J.S. Bach" -> ["Bach"]
// "Johann Sebastian Bach" -> ["Bach"]
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

	// Last word is typically the last name
	// Handle compound last names with lowercase particles: "van", "von", "de", "da", etc.
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
