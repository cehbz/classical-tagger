package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ArrangerCredit checks that arrangements are properly credited (classical.arrangement)
// Arranger should be in track title if arrangement is significant
func (r *Rules) ArrangerCredit(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.arrangement",
		Name:   "Arrangements should credit arranger in track title",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

	var issues []domain.ValidationIssue

	for _, track := range actual.Tracks {
		// Find arranger in artist list
		var arranger *domain.Artist
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleArranger {
				arranger = &artist
				break
			}
		}

		// If there's an arranger, check if mentioned in title
		if arranger != nil {
			title := track.Title
			arrangerName := arranger.Name

			// Check if arranger is mentioned in title
			titleLower := strings.ToLower(title)
			nameLower := strings.ToLower(arrangerName)

			// Extract last name
			nameParts := strings.Fields(arrangerName)
			var lastName string
			if len(nameParts) > 0 {
				lastName = nameParts[len(nameParts)-1]
			}
			lastNameLower := strings.ToLower(lastName)

			// Check for common arrangement indicators
			hasArrangementIndicator := strings.Contains(titleLower, "arr.") ||
				strings.Contains(titleLower, "arranged") ||
				strings.Contains(titleLower, "transcription") ||
				strings.Contains(titleLower, "transcribed")

			// If title indicates arrangement but doesn't mention arranger
			if hasArrangementIndicator && !strings.Contains(titleLower, nameLower) && !strings.Contains(titleLower, lastNameLower) {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelInfo,
					Track: track.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s: Arrangement by '%s' could be credited in title (e.g., 'arr. %s')",
						formatTrackNumber(track), arrangerName, lastName),
				})
			}

			// If no arrangement indicator but has arranger, suggest adding it
			if !hasArrangementIndicator {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelInfo,
					Track: track.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s: Consider adding arrangement credit in title (e.g., 'arr. %s')",
						formatTrackNumber(track), lastName),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
