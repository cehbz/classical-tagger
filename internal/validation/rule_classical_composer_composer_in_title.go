package validation

import (
	"fmt"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// ComposerNotInTitle checks that the composer's surname doesn't appear in track titles (classical.composer rule)
func (r *Rules) ComposerNotInTitle(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "classical.composer",
		name:   "Composer not in track title",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	var issues []domain.ValidationIssue
	
	for _, track := range actual.Tracks() {
		// Find the composer artist
		var composer *domain.Artist
		for _, artist := range track.Artists() {
			if artist.Role() == domain.RoleComposer {
				composer = &artist
				break
			}
		}
		
		if composer == nil {
			// No composer found - different rule handles this
			continue
		}
		
		composerLast := lastName(composer.Name())
		if strings.Contains(track.Title(), composerLast) {
			issues = append(issues, domain.NewIssue(
				domain.LevelError,
				track.Track(),
				meta.id,
				fmt.Sprintf("Composer surname %q found in track title", composerLast),
			))
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// lastName extracts the last name from a full name
func lastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return fullName
	}
	return parts[len(parts)-1]
}