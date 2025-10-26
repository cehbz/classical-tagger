package validation

import (
	"fmt"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// separatorPatterns are indicators of combined/multiple values in a single tag
var separatorPatterns = []string{";", " / ", " & ", ", ", " and "}

// NoCombinedTags checks that tags don't combine multiple values (rule 2.3.18.3)
// Each artist/performer should have separate tag entries
func (r *Rules) NoCombinedTags(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.18.3",
		name:   "No combined tags - use separate entries for multiple artists",
		level:  domain.LevelWarning,
		weight: 0.5,
	}
	
	var issues []domain.ValidationIssue
	
	// Check track titles for combined info that should be separate
	for _, track := range actual.Tracks() {
		title := track.Title()
		
		// Check for multiple works in title (should be separate tracks)
		// Pattern: "Work 1 / Work 2" or "Work 1; Work 2"
		for _, sep := range []string{" / ", "; "} {
			if strings.Contains(title, sep) {
				// Check if this looks like multiple works
				parts := strings.Split(title, sep)
				if len(parts) >= 2 && len(parts[0]) > 10 && len(parts[1]) > 10 {
					issues = append(issues, domain.NewIssue(
						domain.LevelInfo,
						track.Track(),
						meta.id,
						fmt.Sprintf("Track %s: Title may contain multiple works '%s' (consider separate tracks)",
							formatTrackNumber(track), title),
					))
					break
				}
			}
		}
		
		// Check artists for combined names
		// Note: The domain model already handles multiple artists as separate entries
		// This check is for cases where a single artist entry contains multiple names
		for _, artist := range track.Artists() {
			name := artist.Name()
			
			// Check for obvious combined names
			for _, sep := range separatorPatterns {
				if strings.Contains(name, sep) {
					// Some exceptions are valid:
					// - "Orchestra of the Age of Enlightenment" (has " of ", " the ")
					// - "London Symphony Orchestra and Chorus" (ensemble names can have "and")
					// - Compound last names: "Mendelssohn-Bartholdy"
					
					// Check if this looks like multiple people
					if isMultipleArtists(name, sep) {
						issues = append(issues, domain.NewIssue(
							domain.LevelWarning,
							track.Track(),
							meta.id,
							fmt.Sprintf("Track %s: Artist '%s' may contain multiple names (use separate entries)",
								formatTrackNumber(track), name),
						))
						break
					}
				}
			}
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// isMultipleArtists checks if a name string contains multiple distinct artists
func isMultipleArtists(name, separator string) bool {
	// Skip if it's a known ensemble pattern
	lowerName := strings.ToLower(name)
	
	// Orchestra/Choir names often contain separators
	if strings.Contains(lowerName, "orchestra") ||
		strings.Contains(lowerName, "choir") ||
		strings.Contains(lowerName, "chorus") ||
		strings.Contains(lowerName, "ensemble") ||
		strings.Contains(lowerName, "quartet") ||
		strings.Contains(lowerName, "trio") {
		return false
	}
	
	// Titles in names
	if strings.Contains(lowerName, " of ") ||
		strings.Contains(lowerName, " the ") ||
		strings.Contains(lowerName, " de ") ||
		strings.Contains(lowerName, " la ") {
		return false
	}
	
	// Compound last names with hyphen
	if separator == ", " && strings.Contains(name, "-") {
		return false
	}
	
	// If we have a separator and none of the exceptions apply,
	// it's likely multiple artists
	parts := strings.Split(name, separator)
	if len(parts) >= 2 {
		// Both parts should be substantial (not just initials)
		if len(strings.TrimSpace(parts[0])) > 3 && len(strings.TrimSpace(parts[1])) > 3 {
			return true
		}
	}
	
	return false
}
