package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// guestIndicators are phrases that indicate a guest artist
var guestIndicators = []string{
	"feat.", "featuring", "with", "guest",
}

// GuestArtistIdentification checks for proper guest artist handling (classical.guest)
// INFO level - suggests separating guest artists in tags
func (r *Rules) GuestArtistIdentification(actualAlbum, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.guest",
		Name:   "Guest artists should be identified in tags or title",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

	var issues []domain.ValidationIssue

	// Analyze performers across tracks to identify potential guests
	performerCounts := make(map[string]int)
	totalTracks := len(actualAlbum.Tracks)

	if totalTracks == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Count appearances of each performer
	for _, track := range actualAlbum.Tracks {
		for _, artist := range track.Artists {
			// Count soloists and conductors (potential guests)
			if artist.Role == domain.RoleSoloist || artist.Role == domain.RoleConductor {
				performerCounts[artist.Name]++
			}
		}
	}

	// Identify infrequent performers (potential guests)
	for performer, count := range performerCounts {
		// If appears in <30% of tracks (rounding up), might be a guest
		threshold := (totalTracks + 2) / 3
		if count < threshold && totalTracks > 3 {
			// Check if already indicated in title
			hasGuestIndication := false

			for _, track := range actualAlbum.Tracks {
				// Check if this performer is on this track
				hasPerformer := false
				for _, artist := range track.Artists {
					if artist.Name == performer {
						hasPerformer = true
						break
					}
				}

				if hasPerformer {
					titleLower := strings.ToLower(track.Title)
					// Check if title mentions guest status
					for _, indicator := range guestIndicators {
						if strings.Contains(titleLower, indicator) {
							hasGuestIndication = true
							break
						}
					}
				}
			}

			if !hasGuestIndication {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelInfo,
					Track: 0,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Performer '%s' appears infrequently (%d/%d tracks) - consider marking as guest artist",
						performer, count, totalTracks),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
