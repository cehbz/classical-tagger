package validation

import (
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumTitleAccuracy checks that album title matches reference (rule 2.3.6)
// Uses fuzzy matching to allow for minor differences
func (r *Rules) AlbumTitleAccuracy(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.6",
		name:   "Album title must accurately match reference",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	// Only validate if reference is provided
	if reference == nil {
		return meta.Pass()
	}
	
	var issues []domain.ValidationIssue
	
	actualTitle := actual.Title()
	referenceTitle := reference.Title()
	
	if actualTitle == "" || referenceTitle == "" {
		return meta.Pass() // Will be caught by RequiredTags
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
			issues = append(issues, domain.NewIssue(
				domain.LevelError,
				0,
				meta.id,
				fmt.Sprintf("Album title '%s' does not match reference '%s'",
					actualTitle, referenceTitle),
			))
		} else if distance > 3 {
			// Moderate difference
			issues = append(issues, domain.NewIssue(
				domain.LevelWarning,
				0,
				meta.id,
				fmt.Sprintf("Album title '%s' differs from reference '%s' (minor differences)",
					actualTitle, referenceTitle),
			))
		}
		// distance <= 3 is acceptable (handled by titlesMatch)
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
