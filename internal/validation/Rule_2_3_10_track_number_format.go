package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TrackNumberFormat checks track number format consistency (rule 2.3.10)
// INFO level - suggests using consistent track numbering format
func (r *Rules) TrackNumberFormat(actual, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.10",
		Name:   "Track numbers should use consistent format",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

	var issues []domain.ValidationIssue

	tracks := actual.Tracks
	if len(tracks) == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Check for track numbering consistency
	maxTrack := 0
	for _, track := range tracks {
		if track.Track > maxTrack {
			maxTrack = track.Track
		}
	}

	// Check if all track numbers are present (no gaps)
	trackNums := make(map[int]bool)
	for _, track := range tracks {
		trackNums[track.Track] = true
	}

	var gaps []int
	for i := 1; i <= maxTrack; i++ {
		if !trackNums[i] {
			gaps = append(gaps, i)
		}
	}

	if len(gaps) > 0 {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelInfo,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Track numbering has gaps: %v", gaps),
		})
	}

	// Check for multi-disc numbering consistency
	maxDisc := 0
	for _, track := range tracks {
		if track.Disc > maxDisc {
			maxDisc = track.Disc
		}
	}

	if maxDisc > 1 {
		// Check each disc starts at 1
		for disc := 1; disc <= maxDisc; disc++ {
			hasTrackOne := false
			for _, track := range tracks {
				if track.Disc == disc && track.Track == 1 {
					hasTrackOne = true
					break
				}
			}

			if !hasTrackOne {
				issues = append(issues, domain.ValidationIssue{
					Level:   domain.LevelInfo,
					Track:   0,
					Rule:    meta.ID,
					Message: fmt.Sprintf("Disc %d: Track numbering should start at 1", disc),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
