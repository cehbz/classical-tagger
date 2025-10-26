package validation

import (
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumArtistTag checks for album artist tag presence (rule 2.3.9)
// INFO level - recommends using album artist tag for consistency
func (r *Rules) AlbumArtistTag(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.9",
		name:   "Album Artist tag recommended for consistent grouping",
		level:  domain.LevelInfo,
		weight: 0.1,
	}
	
	var issues []domain.ValidationIssue
	
	// Check if album has a consistent primary artist across tracks
	primaryArtists := make(map[string]int)
	totalTracks := len(actual.Tracks())
	
	if totalTracks == 0 {
		return meta.Pass()
	}
	
	// Count primary performers/ensembles across tracks
	for _, track := range actual.Tracks() {
		for _, artist := range track.Artists() {
			// Count non-composers (performers, ensembles, conductors)
			if artist.Role() != domain.RoleComposer && artist.Role() != domain.RoleArranger {
				primaryArtists[artist.Name()]++
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
	
	// If there's a dominant performer, suggest album artist tag
	if dominantArtist != "" && maxCount > totalTracks/2 {
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			fmt.Sprintf("Consider using Album Artist tag with '%s' (appears in %d/%d tracks)",
				dominantArtist, maxCount, totalTracks),
		))
	}
	
	// Check for various artists case
	composers := make(map[string]bool)
	for _, track := range actual.Tracks() {
		composer := getComposer(track.Artists())
		if composer != "" {
			composers[composer] = true
		}
	}
	
	// If multiple composers, suggest "Various Artists"
	if len(composers) > 3 {
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			fmt.Sprintf("Multiple composers (%d) detected - consider Album Artist tag 'Various Artists'",
				len(composers)),
		))
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
