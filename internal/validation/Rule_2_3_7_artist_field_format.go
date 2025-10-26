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
func (r *Rules) ArtistFieldFormat(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.7",
		name:   "Artist field format must follow conventions",
		level:  domain.LevelWarning,
		weight: 0.5,
	}
	
	var issues []domain.ValidationIssue
	
	for _, track := range actual.Tracks() {
		artists := track.Artists()
		if len(artists) == 0 {
			continue
		}
		
		// Check that composer is not the only artist
		// (will be caught by other rules but good to check)
		hasPerformer := false
		var composer *domain.Artist
		
		for _, artist := range artists {
			if artist.Role() == domain.RoleComposer {
				composer = &artist
			} else if artist.Role() != domain.RoleArranger {
				hasPerformer = true
			}
		}
		
		if !hasPerformer && composer != nil {
			issues = append(issues, domain.NewIssue(
				domain.LevelWarning,
				track.Track(),
				meta.id,
				fmt.Sprintf("Track %s: Artist field should contain performers, not just composer '%s'",
					formatTrackNumber(track), composer.Name()),
			))
		}
		
		// Check for proper formatting if reference is provided
		if reference != nil {
			refTracks := reference.Tracks()
			refTrackMap := make(map[string]*domain.Track)
			for _, rt := range refTracks {
				key := fmt.Sprintf("%d-%d", rt.Disc(), rt.Track())
				refTrackMap[key] = rt
			}
			
			key := fmt.Sprintf("%d-%d", track.Disc(), track.Track())
			if refTrack, exists := refTrackMap[key]; exists {
				// Compare performer lists
				actualPerformers := getPerformers(track.Artists())
				refPerformers := getPerformers(refTrack.Artists())
				
				if len(actualPerformers) != len(refPerformers) {
					issues = append(issues, domain.NewIssue(
						domain.LevelInfo,
						track.Track(),
						meta.id,
						fmt.Sprintf("Track %s: Number of performers (%d) differs from reference (%d)",
							formatTrackNumber(track), len(actualPerformers), len(refPerformers)),
					))
				}
			}
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// getPerformers extracts performer names (non-composer, non-arranger)
func getPerformers(artists []domain.Artist) []string {
	var performers []string
	for _, artist := range artists {
		if artist.Role() != domain.RoleComposer && artist.Role() != domain.RoleArranger {
			performers = append(performers, artist.Name())
		}
	}
	return performers
}
