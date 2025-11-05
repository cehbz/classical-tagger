package validation

import (
	"fmt"
	"sort"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// FilenameSortingOrder checks that filenames sort alphabetically into playback order (rule 2.3.14)
func (r *Rules) FilenameSortingOrder(actual, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.14",
		Name:   "Filenames sort alphabetically into playback order",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue
	tracks := actual.Tracks()

	if len(tracks) <= 1 {
		return RuleResult{Meta: meta, Issues: nil} // Single track or empty, nothing to sort
	}

	// Group tracks by disc
	discTracks := make(map[int][]*domain.Track)
	for _, track := range tracks {
		disc := track.Disc
		discTracks[disc] = append(discTracks[disc], track)
	}

	// Check each disc separately
	for disc, discTrackList := range discTracks {
		if len(discTrackList) <= 1 {
			continue
		}

		// Create a copy for sorting
		sortedTracks := make([]*domain.Track, len(discTrackList))
		copy(sortedTracks, discTrackList)

		// Sort by raw filename (exact bytewise order)
		sort.Slice(sortedTracks, func(i, j int) bool {
			return sortedTracks[i].File.Path < sortedTracks[j].File.Path
		})

		// Build canonical order (by track number)
		canonical := make([]*domain.Track, len(discTrackList))
		copy(canonical, discTrackList)
		sort.Slice(canonical, func(i, j int) bool { return canonical[i].Track < canonical[j].Track })

		// Compare sequences; report only the first mismatch overall
		mismatchReported := false
		for i := range canonical {
			if sortedTracks[i] != canonical[i] {
				a := sortedTracks[i]
				b := canonical[i]
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelError,
					Track: b.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Disc %d: Filename sorting differs at position %d: got '%s' (track %d), expected '%s' (track %d)",
						disc, i+1, a.File.Path, a.Track, b.File.Path, b.Track),
				})
				mismatchReported = true
				break
			}
		}
		if mismatchReported {
			// No need to continue collecting multiple issues across discs per requirement
			break
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
