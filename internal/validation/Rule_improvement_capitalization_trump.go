package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// CapitalizationTrump checks if proper capitalization alone justifies replacing torrent
// This is a specialized improvement rule focused on capitalization
func (r *Rules) CapitalizationTrump(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "improvement.capitalization",
		Name:   "Proper capitalization can trump existing torrent if significantly better",
		Level:  domain.LevelInfo,
		Weight: 2.0,
	}

	// Can only evaluate if reference exists
	if reference == nil {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	// Count capitalization issues in actual vs reference
	actualIssues := r.countCapitalizationIssues(actual)
	refIssues := r.countCapitalizationIssues(reference)

	// Calculate improvement
	improvement := refIssues - actualIssues

	if improvement <= 0 {
		// No improvement or worse
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Capitalization quality: actual has %d issues vs reference %d issues (no improvement)",
				actualIssues, refIssues),
		})
	} else if improvement < 5 {
		// Minor improvement
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Capitalization improvement: %d fewer issues than reference (minor improvement)",
				improvement),
		})
	} else if improvement < 10 {
		// Moderate improvement
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Capitalization improvement: %d fewer issues than reference (moderate improvement - may justify replacement)",
				improvement),
		})
	} else {
		// Significant improvement
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Capitalization improvement: %d fewer issues than reference (significant improvement - justifies replacement)",
				improvement),
		})
	}

	// Always pass - this is informational
	if len(issues) == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// countCapitalizationIssues counts all capitalization problems in an album
func (r *Rules) countCapitalizationIssues(album *domain.Album) int {
	count := 0

	// Count album title capitalization
	if checkCapitalization(album.Title) != "" {
		count++
	}

	// Check all track titles and artists
	for _, track := range album.Tracks {
		// Track title
		if checkCapitalization(track.Title) != "" {
			count++
		}
	}

	return count
}
