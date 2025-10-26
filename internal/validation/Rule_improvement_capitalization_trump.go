package validation

import (
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// CapitalizationTrump checks if proper capitalization alone justifies replacing torrent
// This is a specialized improvement rule focused on capitalization
func (r *Rules) CapitalizationTrump(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "improvement.capitalization",
		name:   "Proper capitalization can trump existing torrent if significantly better",
		level:  domain.LevelInfo,
		weight: 2.0,
	}
	
	// Can only evaluate if reference exists
	if reference == nil {
		return meta.Pass()
	}
	
	var issues []domain.ValidationIssue
	
	// Count capitalization issues in actual vs reference
	actualIssues := r.countCapitalizationIssues(actual)
	refIssues := r.countCapitalizationIssues(reference)
	
	// Calculate improvement
	improvement := refIssues - actualIssues
	
	if improvement <= 0 {
		// No improvement or worse
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			fmt.Sprintf("Capitalization quality: actual has %d issues vs reference %d issues (no improvement)",
				actualIssues, refIssues),
		))
	} else if improvement < 5 {
		// Minor improvement
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			fmt.Sprintf("Capitalization improvement: %d fewer issues than reference (minor improvement)",
				improvement),
		))
	} else if improvement < 10 {
		// Moderate improvement
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			fmt.Sprintf("Capitalization improvement: %d fewer issues than reference (moderate improvement - may justify replacement)",
				improvement),
		))
	} else {
		// Significant improvement
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			fmt.Sprintf("Capitalization improvement: %d fewer issues than reference (significant improvement - justifies replacement)",
				improvement),
		))
	}
	
	// Always pass - this is informational
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// countCapitalizationIssues counts all capitalization problems in an album
func (r *Rules) countCapitalizationIssues(album *domain.Album) int {
	count := 0
	
	// Check album title
	if checkCapitalization(album.Title()) != "" {
		count++
	}
	
	// Check all track titles and artists
	for _, track := range album.Tracks() {
		// Track title
		if checkCapitalization(track.Title()) != "" {
			count++
		}
		
		// Artist names
		for _, artist := range track.Artists() {
			if checkCapitalization(artist.Name()) != "" {
				count++
			}
		}
	}
	
	return count
}
