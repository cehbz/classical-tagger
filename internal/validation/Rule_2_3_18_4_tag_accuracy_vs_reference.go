package validation

import (
	"fmt"
	"regexp"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TagAccuracyVsReference checks all tags against reference data (rule 2.3.18.4)
// Comprehensive validation of all tag fields
func (r *Rules) TagAccuracyVsReference(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.4",
		Name:   "All tags must accurately match reference data",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	// Only validate if reference is provided
	if reference == nil {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	// Validate year
	if actual.OriginalYear != 0 && reference.OriginalYear != 0 {
		if actual.OriginalYear != reference.OriginalYear {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelWarning,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Album year %d doesn't match reference %d",
					actual.OriginalYear, reference.OriginalYear),
			})
		}
	}

	// Validate track-level tags
	actualTracks := actual.Tracks
	refTracks := reference.Tracks

	// Create a map for easier track matching
	refTrackMap := make(map[string]*domain.Track)
	for _, refTrack := range refTracks {
		key := fmt.Sprintf("%d-%d", refTrack.Disc, refTrack.Track)
		refTrackMap[key] = refTrack
	}

	for _, actualTrack := range actualTracks {
		key := fmt.Sprintf("%d-%d", actualTrack.Disc, actualTrack.Track)
		refTrack, exists := refTrackMap[key]

		if !exists {
			continue // No reference track to compare
		}

		// Compare track titles
		normActual := normalizeTitle(actualTrack.Title)
		normRef := normalizeTitle(refTrack.Title)
		if normActual != normRef {
			// Calculate severity based on difference
			distance := levenshteinDistance(normActual, normRef)

			level := domain.LevelError
			// If titles differ only by work number (e.g., No. 6 vs No. 5), treat as error
			if differentWorkNumber(normActual, normRef) {
				level = domain.LevelError
			} else {
				if distance <= 3 {
					level = domain.LevelInfo // Minor difference
				} else if distance <= 10 {
					level = domain.LevelWarning // Moderate difference
				}
			}

			issues = append(issues, domain.ValidationIssue{
				Level: level,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Title '%s' doesn't match reference '%s'",
					formatTrackNumber(actualTrack), actualTrack.Title, refTrack.Title),
			})
		}

		// Compare composers
		actualComposer := getComposer(actualTrack.Artists)
		refComposer := getComposer(refTrack.Artists)

		if actualComposer != "" && refComposer != "" && actualComposer != refComposer {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Composer '%s' doesn't match reference '%s'",
					formatTrackNumber(actualTrack), actualComposer, refComposer),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

var workNoRe = regexp.MustCompile(`(?i)\bno\.?\s*(\d+)`)

func workNumber(title string) string {
	m := workNoRe.FindStringSubmatch(title)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

func differentWorkNumber(a, b string) bool {
	na := workNumber(a)
	nb := workNumber(b)
	return na != "" && nb != "" && na != nb
}

// getComposer extracts composer name from artist list
func getComposer(artists []domain.Artist) string {
	for _, artist := range artists {
		if artist.Role == domain.RoleComposer {
			return artist.Name
		}
	}
	return ""
}
