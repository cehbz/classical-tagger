package validation

import (
	"fmt"
	"regexp"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// discNumberPattern matches disc numbers in various formats
// Examples: "Disc 1", "CD 2", "Volume 3", "Vol. 1"
var discNumberPattern = regexp.MustCompile(`(?i)(disc|cd|disk|volume|vol\.?)\s*\d+`)

// NoDiscNumbersInAlbumTag checks that disc numbers aren't in album title (rule 2.3.18.3.3)
// Disc numbers should be in the DISC tag, not the album title
func (r *Rules) NoDiscNumbersInAlbumTag(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.18.3.3",
		name:   "Disc numbers should not be in album title",
		level:  domain.LevelWarning,
		weight: 0.5,
	}
	
	var issues []domain.ValidationIssue
	
	albumTitle := actual.Title()
	if albumTitle == "" {
		return meta.Pass()
	}
	
	// Check if album title contains disc number indicators
	if discNumberPattern.MatchString(albumTitle) {
		// Check if this is a legitimate multi-volume set name
		// "Complete Works, Volume 1" is OK for a series
		// "Symphony No. 5 - Disc 1" is NOT OK
		
		if !isLegitimateVolumeTitle(albumTitle) {
			issues = append(issues, domain.NewIssue(
				domain.LevelWarning,
				0,
				meta.id,
				fmt.Sprintf("Album title '%s' contains disc number (should use DISC tag instead)", albumTitle),
			))
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// isLegitimateVolumeTitle checks if the title legitimately contains volume/disc numbers
// as part of a series name rather than as disc metadata
func isLegitimateVolumeTitle(title string) bool {
	lowerTitle := strings.ToLower(title)
	
	// These patterns indicate legitimate volume series:
	// - "Complete Works, Volume 1-5"
	// - "Collected Recordings, Vol. 1"
	// - "The Decca Recordings, Volume 2"
	
	legitimatePatterns := []string{
		"complete works",
		"collected",
		"recordings",
		"anthology",
		"collection",
		"edition",
		"series",
	}
	
	for _, pattern := range legitimatePatterns {
		if strings.Contains(lowerTitle, pattern) {
			return true
		}
	}
	
	// Check if it's a range: "Volume 1-5" or "Discs 1-3"
	if strings.Contains(lowerTitle, "-") {
		parts := strings.Split(lowerTitle, "-")
		for _, part := range parts {
			// If there's a digit before and after the dash near volume/disc, it's a range
			if regexp.MustCompile(`\d+\s*$`).MatchString(strings.TrimSpace(part)) {
				return true
			}
		}
	}
	
	return false
}
