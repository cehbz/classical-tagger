package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// composerNamePattern checks for first name or initial
// Accepts: "Johann Sebastian Bach", "J.S. Bach", "J. S. Bach"
// Rejects: "Bach" (ambiguous)
var composerNamePattern = regexp.MustCompile(`^[A-Z]\S*[\s\.]+\S+|^\S+\s+\S+`)

// ComposerTagRequired checks that composer tag is present and uniquely identifiable (classical.composer)
func (r *Rules) ComposerTagRequired(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.composer",
		Name:   "Composer tag required with identifiable name",
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
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   track.Track,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Track %s: Composer tag is missing", formatTrackNumber(track)),
			})
			continue
		}

		// Check that composer name is uniquely identifiable
		// Must have at least first name or initial, not just last name
		composerName := composer.Name

		// Check for ambiguous single-word names
		if !strings.Contains(composerName, " ") && !strings.Contains(composerName, ".") {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Composer name '%s' is not uniquely identifiable (needs first name or initial)",
					formatTrackNumber(track), composerName),
			})
			continue
		}

		// Additional check: name should match pattern for completeness
		if !composerNamePattern.MatchString(composerName) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelWarning,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Composer name '%s' may not be in standard format (e.g., 'Johann Sebastian Bach' or 'J.S. Bach')",
					formatTrackNumber(track), composerName),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
