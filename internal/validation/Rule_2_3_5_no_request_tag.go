package validation

import (
	"fmt"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// NoRequestTagInTitle checks that album title doesn't contain [REQ] tag (rule 2.3.5)
// [REQ] is used in requests but should be removed from actual torrents
func (r *Rules) NoRequestTagInTitle(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.5",
		name:   "Album title must not contain [REQ] tag",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	var issues []domain.ValidationIssue
	
	albumTitle := actual.Title()
	if albumTitle == "" {
		return meta.Pass()
	}
	
	// Check for [REQ] tag (case insensitive)
	albumTitleUpper := strings.ToUpper(albumTitle)
	
	if strings.Contains(albumTitleUpper, "[REQ]") {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			0,
			meta.id,
			fmt.Sprintf("Album title '%s' contains [REQ] tag (must be removed)", albumTitle),
		))
	}
	
	// Also check for common variants
	variants := []string{"[REQUEST]", "[REQUESTED]", "(REQ)", "(REQUEST)"}
	for _, variant := range variants {
		if strings.Contains(albumTitleUpper, variant) {
			issues = append(issues, domain.NewIssue(
				domain.LevelWarning,
				0,
				meta.id,
				fmt.Sprintf("Album title '%s' contains request indicator '%s' (should be removed)",
					albumTitle, variant),
			))
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}