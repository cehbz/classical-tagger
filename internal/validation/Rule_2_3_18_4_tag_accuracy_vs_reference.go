package validation

import (
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// TagAccuracyVsReference checks all tags against reference data (rule 2.3.18.4)
// Comprehensive validation of all tag fields
func (r *Rules) TagAccuracyVsReference(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.18.4",
		name:   "All tags must accurately match reference data",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	// Only validate if reference is provided
	if reference == nil {
		return meta.Pass()
	}
	
	var issues []domain.ValidationIssue
	
	// Validate year
	if actual.OriginalYear() != 0 && reference.OriginalYear() != 0 {
		if actual.OriginalYear() != reference.OriginalYear() {
			issues = append(issues, domain.NewIssue(
				domain.LevelWarning,
				0,
				meta.id,
				fmt.Sprintf("Album year %d doesn't match reference %d",
					actual.OriginalYear(), reference.OriginalYear()),
			))
		}
	}
	
	// Validate track-level tags
	actualTracks := actual.Tracks()
	refTracks := reference.Tracks()
	
	// Create a map for easier track matching
	refTrackMap := make(map[string]*domain.Track)
	for _, refTrack := range refTracks {
		key := fmt.Sprintf("%d-%d", refTrack.Disc(), refTrack.Track())
		refTrackMap[key] = refTrack
	}
	
	for _, actualTrack := range actualTracks {
		key := fmt.Sprintf("%d-%d", actualTrack.Disc(), actualTrack.Track())
		refTrack, exists := refTrackMap[key]
		
		if !exists {
			continue // No reference track to compare
		}
		
		// Compare track titles
		if !titlesMatch(
			normalizeTitle(actualTrack.Title()),
			normalizeTitle(refTrack.Title()),
		) {
			// Calculate severity based on difference
			distance := levenshteinDistance(
				normalizeTitle(actualTrack.Title()),
				normalizeTitle(refTrack.Title()),
			)
			
			level := domain.LevelError
			if distance <= 3 {
				level = domain.LevelInfo // Minor difference
			} else if distance <= 10 {
				level = domain.LevelWarning // Moderate difference
			}
			
			issues = append(issues, domain.NewIssue(
				level,
				actualTrack.Track(),
				meta.id,
				fmt.Sprintf("Track %s: Title '%s' doesn't match reference '%s'",
					formatTrackNumber(actualTrack), actualTrack.Title(), refTrack.Title()),
			))
		}
		
		// Compare composers
		actualComposer := getComposer(actualTrack.Artists())
		refComposer := getComposer(refTrack.Artists())
		
		if actualComposer != "" && refComposer != "" && actualComposer != refComposer {
			issues = append(issues, domain.NewIssue(
				domain.LevelError,
				actualTrack.Track(),
				meta.id,
				fmt.Sprintf("Track %s: Composer '%s' doesn't match reference '%s'",
					formatTrackNumber(actualTrack), actualComposer, refComposer),
			))
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// getComposer extracts composer name from artist list
func getComposer(artists []domain.Artist) string {
	for _, artist := range artists {
		if artist.Role() == domain.RoleComposer {
			return artist.Name()
		}
	}
	return ""
}
