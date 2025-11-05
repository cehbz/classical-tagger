package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// RequiredTags checks that all required tags are present (rule 2.3.16.4)
// Required: Artist, Album, Title, Track Number
// Optional but encouraged: Year
func (r *Rules) RequiredAlbumTags(actualTorrent, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.16.4-album",
		Name:   "Required tags present",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check album-level tags

	// Album title (required)
	if actualTorrent.Title == "" {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: "Album title tag is missing",
		})
	}

	// Year (optional but strongly encouraged)
	if actualTorrent.OriginalYear == 0 {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: "Year tag is missing (strongly recommended)",
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// RequiredTags checks that all required tags are present (rule 2.3.16.4)
// Required: Artist, Album, Title, Track Number
// Optional but encouraged: Year
func (r *Rules) RequiredTrackTags(actualTrack, _ *domain.Track, actualTorrent, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.16.4-track",
		Name:   "Required tags present",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check track-level tags
	trackNum := actualTrack.Track

	// Title (required)
	if actualTrack.Title == "" {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   trackNum,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Track %s: Title tag is missing", formatTrackNumber(actualTrack)),
		})
	}

	// Track Number (required) - checked implicitly by domain model
	// The domain.NewTrack() requires track number, so this is always present

	// Artist (required)
	artists := actualTrack.Artists
	if len(artists) == 0 {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   trackNum,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Track %s: Artist tag is missing", formatTrackNumber(actualTrack)),
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
			// Enforce when single-track album or when the track title itself is missing
			// (indicates an incomplete tagging situation). Multi-track composer-only
			// entries with proper titles are handled by performer-format rules.
			if actualTorrent == nil || len(actualTorrent.Tracks()) <= 1 || strings.TrimSpace(actualTrack.Title) == "" {
				issues = append(issues, domain.ValidationIssue{
					Level:   domain.LevelError,
					Track:   trackNum,
					Rule:    meta.ID,
					Message: fmt.Sprintf("Track %s: Artist tag has no performers (only composer)", formatTrackNumber(actualTrack)),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
