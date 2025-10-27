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
		ID:     "2.3.5",
		Name:   "Album title must not contain [REQ] tag",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	albumTitle := actual.Title
	if albumTitle == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Check for [REQ] tag (case insensitive)
	albumTitleUpper := strings.ToUpper(albumTitle)

	if strings.Contains(albumTitleUpper, "[REQ]") {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title '%s' contains [REQ] tag (must be removed)", albumTitle),
		})
	}

	// Also check for common variants
	variants := []string{"[REQUEST]", "[REQUESTED]", "(REQ)", "(REQUEST)"}
	for _, variant := range variants {
		if strings.Contains(albumTitleUpper, variant) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelWarning,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Album title '%s' contains request indicator '%s' (should be removed)",
					albumTitle, variant),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
