package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// artistFieldPattern validates the Artist field format per rule 2.3.7
// Should follow format: "Soloist(s); Ensemble; Conductor"
var artistFieldSeparator = "; "

// ArtistFieldFormat checks artist field formatting (rule 2.3.7)
// For classical: performers listed, not composer
// Format should be consistent with semicolon separation if multiple
func (r *Rules) ArtistFieldFormat(actualTrack, refTrack *domain.Track, actualTorrent, refTorrent *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.7-format",
		Name:   "Artist field format must follow conventions",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}
	if actualTrack == nil || len(actualTrack.Artists) == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	artists := actualTrack.Artists
	// Check that composer is not the only artist
	// (will be caught by other rules but good to check)
	hasPerformer := false
	var composer *domain.Artist

	for _, artist := range artists {
		if artist.Role == domain.RoleComposer {
			composer = &artist
		} else if artist.Role != domain.RoleArranger {
			hasPerformer = true
		}
	}

	if !hasPerformer && composer != nil {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelWarning,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Artist field should contain performers, not just composer '%s'",
				formatTrackNumber(actualTrack), composer.Name),
		})
	}

	if refTrack == nil {
		return RuleResult{Meta: meta, Issues: issues}
	}

	// compare disc and track numbers
	if refTrack.Disc != actualTrack.Disc || refTrack.Track != actualTrack.Track {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Reference track does not match actual track (disc %d, track %d)",
				formatTrackNumber(actualTrack), refTrack.Disc, refTrack.Track),
		})
	}

	// Compare performer lists
	actualPerformers := actualTrack.Performers()
	refPerformers := refTrack.Performers()

	if len(actualPerformers) != len(refPerformers) {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Number of performers (%d) differs from reference (%d)",
				formatTrackNumber(actualTrack), len(actualPerformers), len(refPerformers)),
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}
