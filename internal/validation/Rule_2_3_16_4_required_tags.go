package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// RequiredTags checks that all required tags are present (rule 2.3.16.4)
// Required: Artist, Album, Title, Track Number
// Optional but encouraged: Year
func (r *Rules) RequiredTags(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.16.4",
		Name:   "Required tags present",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Guard against nil album
	if actual == nil {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: "Album title tag is missing",
		})
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: "Year tag is missing (strongly recommended)",
		})
		if len(issues) == 0 {
			return RuleResult{Meta: meta, Issues: nil}
		}
		return RuleResult{Meta: meta, Issues: issues}
	}

	// Check album-level tags

	// Album title (required)
	if actual.Title == "" {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: "Album title tag is missing",
		})
	}

	// Year (optional but strongly encouraged)
	if actual.OriginalYear == 0 {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: "Year tag is missing (strongly recommended)",
		})
	}

	// Check track-level tags
	for _, track := range actual.Tracks {
		trackNum := track.Track

		// Title (required)
		if track.Title == "" {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   trackNum,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Track %s: Title tag is missing", formatTrackNumber(track)),
			})
		}

		// Track Number (required) - checked implicitly by domain model
		// The domain.NewTrack() requires track number, so this is always present

		// Artist (required)
		artists := track.Artists
		if len(artists) == 0 {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   trackNum,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Track %s: Artist tag is missing", formatTrackNumber(track)),
			})
		} else {
			// Check if there's at least one performer (non-composer)
			hasPerformer := false
			for _, artist := range artists {
				if artist.Role != domain.RoleComposer {
					hasPerformer = true
					break
				}
			}
			if !hasPerformer {
				issues = append(issues, domain.ValidationIssue{
					Level:   domain.LevelError,
					Track:   trackNum,
					Rule:    meta.ID,
					Message: fmt.Sprintf("Track %s: Artist tag has no performers (only composer)", formatTrackNumber(track)),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
