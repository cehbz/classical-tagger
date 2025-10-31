package validation

import (
	"fmt"
    "strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumArtistTag checks for album artist tag presence (rule 2.3.7)
// INFO level - recommends using album artist tag for consistency
func (r *Rules) AlbumArtistTag(actualAlbum, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.7-present",
		Name:   "Album Artist tag recommended for consistent grouping",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

    var issues []domain.ValidationIssue

	// Check if album has a consistent primary artist across tracks
	primaryArtists := make(map[string]int)
	totalTracks := len(actualAlbum.Tracks)

	if totalTracks == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}

    // If title indicates Various Artists, emit info
    if strings.Contains(strings.ToLower(actualAlbum.Title), "various artists") {
        issues = append(issues, domain.ValidationIssue{
            Level: domain.LevelInfo,
            Track: 0,
            Rule:  meta.ID,
            Message: "Album likely 'Various Artists' - consider setting Album Artist accordingly",
        })
        return RuleResult{Meta: meta, Issues: issues}
    }

    // Count primary performers/ensembles across tracks (exclude composers/arrangers)
	for _, track := range actualAlbum.Tracks {
		for _, artist := range track.Artists {
			// Count non-composers (performers, ensembles, conductors)
			if artist.Role != domain.RoleComposer && artist.Role != domain.RoleArranger {
				primaryArtists[artist.Name]++
			}
		}
	}

	// Check if there's a dominant performer (appears in >50% of tracks)
	var dominantArtist string
	maxCount := 0

	for artist, count := range primaryArtists {
		if count > maxCount {
			maxCount = count
			dominantArtist = artist
		}
	}

    // Single track: only suggest when the primary credit is ensemble/conductor (not soloist)
    if totalTracks == 1 {
        hasSoloist := false
        hasEnsembleOrConductor := false
        for _, a := range actualAlbum.Tracks[0].Artists {
            if a.Role == domain.RoleSoloist { hasSoloist = true }
            if a.Role == domain.RoleEnsemble || a.Role == domain.RoleConductor { hasEnsembleOrConductor = true }
        }
        if hasEnsembleOrConductor && !hasSoloist {
            issues = append(issues, domain.ValidationIssue{
                Level: domain.LevelInfo,
                Track: 0,
                Rule:  meta.ID,
                Message: "Consider using Album Artist tag for ensemble/conductor",
            })
        }
        return RuleResult{Meta: meta, Issues: issues}
    }

    // If there are exactly two distinct composers, do not suggest dominance
    composers := make(map[string]bool)
    for _, track := range actualAlbum.Tracks {
        c := getComposer(track.Artists)
        if c != "" { composers[c] = true }
    }
    if len(composers) == 2 {
        return RuleResult{Meta: meta, Issues: issues}
    }

    // If there's a dominant performer across multiple tracks, suggest album artist tag
    if dominantArtist != "" && maxCount > totalTracks/2 {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Consider using Album Artist tag with '%s' (appears in %d/%d tracks)",
				dominantArtist, maxCount, totalTracks),
		})
	}

    // If multiple composers, suggest "Various Artists"
    if len(composers) > 3 {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelInfo,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Multiple composers (%d) detected - consider Album Artist tag 'Various Artists'",
				len(composers)),
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}
