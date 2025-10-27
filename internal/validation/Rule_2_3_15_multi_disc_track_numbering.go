package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// MultiDiscTrackNumbering checks that multi-disc releases number tracks correctly (rule 2.3.15)
// Each disc should start at track 1
func (r *Rules) MultiDiscTrackNumbering(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.15",
		Name:   "Multi-disc track numbering starts at 1 for each disc",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Group tracks by disc
	discTracks := make(map[int][]*domain.Track)
	maxDisc := 1

	for _, track := range actual.Tracks {
		disc := track.Disc
		if disc > maxDisc {
			maxDisc = disc
		}
		discTracks[disc] = append(discTracks[disc], track)
	}

	// Single disc - no special numbering rules
	if maxDisc == 1 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Multi-disc: check each disc
	for disc := 1; disc <= maxDisc; disc++ {
		tracks := discTracks[disc]
		if len(tracks) == 0 {
			// Missing disc in sequence
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Multi-disc release missing disc %d", disc),
			})
			continue
		}

		// Check if tracks start at 1
		hasTrackOne := false
		lowestTrack := 999999

		for _, track := range tracks {
			trackNum := track.Track
			if trackNum == 1 {
				hasTrackOne = true
			}
			if trackNum < lowestTrack {
				lowestTrack = trackNum
			}
		}

		if !hasTrackOne {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Disc %d: Track numbering must start at 1 (lowest track is %d)",
					disc, lowestTrack),
			})
		}

		// Check for sequential numbering (no gaps)
		trackNums := make(map[int]bool)
		for _, track := range tracks {
			trackNums[track.Track] = true
		}

		expectedCount := len(tracks)
		for i := 1; i <= expectedCount; i++ {
			if !trackNums[i] {
				issues = append(issues, domain.ValidationIssue{
					Level:   domain.LevelWarning,
					Track:   0,
					Rule:    meta.ID,
					Message: fmt.Sprintf("Disc %d: Gap in track numbering at track %d", disc, i),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
