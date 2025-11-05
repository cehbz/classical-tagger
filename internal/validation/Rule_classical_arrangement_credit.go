package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ArrangerCredit checks that arrangements are properly credited (classical.arrangement)
// Arranger should be in track title if arrangement is significant
func (r *Rules) ArrangerCredit(actualTrack, _ *domain.Track, _, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.arrangement",
		Name:   "Arrangements should credit arranger in track title",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

	var issues []domain.ValidationIssue

	// Find arranger in artist list
	var arranger *domain.Artist
	for _, artist := range actualTrack.Artists {
		if artist.Role == domain.RoleArranger {
			arranger = &artist
			break
		}
	}

	// If there's an arranger, check if mentioned in title
	if arranger == nil {
		return RuleResult{Meta: meta, Issues: nil}
	}
	
	title := actualTrack.Title
	arrangerName := arranger.Name

	// Check if arranger is mentioned in title
	titleLower := strings.ToLower(title)
	nameLower := strings.ToLower(arrangerName)

	// Extract last name
	nameParts := strings.Fields(nameLower)
	lastNameLower := nameLower
	if len(nameParts) > 0 {
		lastNameLower = nameParts[len(nameParts)-1]
	}

	// Check for common arrangement indicators
	hasArrangementIndicator := strings.Contains(titleLower, "arr.") ||
		strings.Contains(titleLower, "arranged") ||
		strings.Contains(titleLower, "transcription") ||
		strings.Contains(titleLower, "transcribed")

	// If title indicates arrangement but doesn't mention arranger
	if hasArrangementIndicator && !strings.Contains(titleLower, nameLower) && !strings.Contains(titleLower, lastNameLower) {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Indicates arrangement, but '%s' is not mentioned in title",
				formatTrackNumber(actualTrack), arrangerName),
		})
	}

	// If no arrangement indicator but has arranger, suggest adding it
	if !hasArrangementIndicator {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Consider adding arrangement credit in title (e.g., 'arr. %s')",
				formatTrackNumber(actualTrack), lastNameLower),
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}
