package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumTitleAccuracy checks that album title matches reference (rule 2.3.6)
// Uses fuzzy matching to allow for minor differences
func (r *Rules) AlbumTitleAccuracy(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.6",
		Name:   "Album title must accurately match reference",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	// Only validate if reference is provided
	if reference == nil {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	actualTitle := actual.Title
	referenceTitle := reference.Title

	if actualTitle == "" || referenceTitle == "" {
		return RuleResult{Meta: meta, Issues: nil} // Will be caught by RequiredTags
	}

	// Normalize both titles for comparison
	normalizedActual := normalizeTitle(actualTitle)
	normalizedReference := normalizeTitle(referenceTitle)

	// Check if titles match
	if !titlesMatch(normalizedActual, normalizedReference) {
		// Check edit distance for severity
		distance := levenshteinDistance(normalizedActual, normalizedReference)

		if distance > 10 {
			// Major difference
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Album title '%s' does not match reference '%s'",
					actualTitle, referenceTitle),
			})
		} else if distance > 3 {
			// Moderate difference
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelWarning,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Album title '%s' differs from reference '%s' (minor differences)",
					actualTitle, referenceTitle),
			})
		}
		// distance <= 3 is acceptable (handled by titlesMatch)
	}
	return RuleResult{Meta: meta, Issues: issues}
}
