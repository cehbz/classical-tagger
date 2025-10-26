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
		id:     "2.3.16.4",
		name:   "Required tags present",
		level:  domain.LevelError,
		weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Guard against nil album
	if actual == nil {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			0,
			meta.id,
			"Album title tag is missing",
		))
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0,
			meta.id,
			"Year tag is missing (strongly recommended)",
		))
		if len(issues) == 0 {
			return meta.Pass()
		}
		return meta.Fail(issues...)
	}

	// Check album-level tags

	// Album title (required)
	if actual.Title() == "" {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			0,
			meta.id,
			"Album title tag is missing",
		))
	}

	// Year (optional but strongly encouraged)
	if actual.OriginalYear() == 0 {
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0,
			meta.id,
			"Year tag is missing (strongly recommended)",
		))
	}

	// Check track-level tags
	for _, track := range actual.Tracks() {
		trackNum := track.Track()

		// Title (required)
		if track.Title() == "" {
			issues = append(issues, domain.NewIssue(
				domain.LevelError,
				trackNum,
				meta.id,
				fmt.Sprintf("Track %s: Title tag is missing", formatTrackNumber(track)),
			))
		}

		// Track Number (required) - checked implicitly by domain model
		// The domain.NewTrack() requires track number, so this is always present

		// Artist (required)
		artists := track.Artists()
		if len(artists) == 0 {
			issues = append(issues, domain.NewIssue(
				domain.LevelError,
				trackNum,
				meta.id,
				fmt.Sprintf("Track %s: Artist tag is missing", formatTrackNumber(track)),
			))
		} else {
			// Check if there's at least one performer (non-composer)
			hasPerformer := false
			for _, artist := range artists {
				if artist.Role() != domain.RoleComposer {
					hasPerformer = true
					break
				}
			}
			if !hasPerformer {
				issues = append(issues, domain.NewIssue(
					domain.LevelError,
					trackNum,
					meta.id,
					fmt.Sprintf("Track %s: Artist tag has no performers (only composer)", formatTrackNumber(track)),
				))
			}
		}
	}

	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
