package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ComposerNotInTitle checks that composer name is NOT in track title (classical.track_title)
// The composer should only be in the COMPOSER tag, not repeated in the title
func (r *Rules) ComposerNotInTitle(actualTrack, _ *domain.Track, _, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.track_title",
		Name:   "Composer name not in track title",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Find the composer
	composers := actualTrack.Composers()

	if len(composers) == 0 {
		// No composer to check - this will be caught by ComposerTagRequired rule
		return RuleResult{Meta: meta, Issues: nil}
	}

	title := actualTrack.Title
	if title == "" {
		// Empty title - will be caught by RequiredTags rule
		return RuleResult{Meta: meta, Issues: nil}
	}

	for _, composer := range composers {
		// Detect using base surname (particle-independent), e.g., "Beethoven", "Bach"
		base := baseSurnameFromFullName(composer.Name)
		// Use word boundaries to avoid false positives
		titleLower := strings.ToLower(title)
		baseLower := strings.ToLower(base)

		// Create a word boundary pattern
		// Match: "Bach: Symphony" or "Symphony (Bach)" or "Bach's"
		// Don't match: "Bacharach" (different word)
		patternWith := regexp.MustCompile(`\b` + regexp.QuoteMeta(baseLower) + `\b`)

		if patternWith.MatchString(titleLower) {
			// Check for exceptions: "Variations on a Theme by Brahms" is acceptable
			// These are part of the actual work title, not just composer mention
			if isComposerPartOfWorkTitle(title, base) {
				continue
			}

			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Composer surname '%s' found in track title '%s' (composer should only be in COMPOSER tag)",
					formatTrackNumber(actualTrack), base, title),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
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

// baseSurnameFromFullName returns the sortable base surname (no particles)
// Examples:
//
//	"Ludwig van Beethoven" -> "Beethoven"
//	"Beethoven, Ludwig van" -> "Beethoven"
//	"J.S. Bach" -> "Bach"
func baseSurnameFromFullName(fullName string) string {
	if strings.Contains(fullName, ",") {
		parts := strings.Split(fullName, ",")
		return strings.TrimSpace(parts[0])
	}
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return fullName
	}
	return parts[len(parts)-1]
}
