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
func (r *Rules) NoDiscNumbersInAlbumTag(actual, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.3.3",
		Name:   "Disc numbers should not be in album title",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	albumTitle := actual.Title
	if albumTitle == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Check if album title contains disc number indicators
	if discNumberPattern.MatchString(albumTitle) {
		// Check if this is a legitimate multi-volume set name
		// "Complete Works, Volume 1" is OK for a series
		// "Symphony No. 5 - Disc 1" is NOT OK

		if !isLegitimateVolumeTitle(albumTitle) {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Album title '%s' contains disc number (should use DISC tag instead)", albumTitle),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// isLegitimateVolumeTitle checks if the title legitimately contains volume/disc numbers
// as part of a series name rather than as disc metadata
func isLegitimateVolumeTitle(title string) bool {
	lowerTitle := strings.ToLower(title)

	// These patterns indicate legitimate volume series context
	// (must appear near volume/disc keywords)
	seriesContext := []string{
		"complete works",
		"collected",
		"recordings",
		"anthology",
		"collection",
		"edition",
		"series",
	}

	for _, ctx := range seriesContext {
		if strings.Contains(lowerTitle, ctx) && regexp.MustCompile(`(?i)(volume|vol\.?|disc|cd)s?\s*\d+`).MatchString(lowerTitle) {
			return true
		}
	}

	// Allow explicit numeric ranges only when tied to volume/disc keywords, e.g. "Volume 1-5"
	rangeRe := regexp.MustCompile(`(?i)(volume|vol\.?|disc|cd)s?\s*\d+\s*-\s*\d+`)
	if rangeRe.MatchString(lowerTitle) {
		return true
	}

	return false
}
