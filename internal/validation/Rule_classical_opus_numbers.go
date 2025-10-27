package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// opusPattern matches opus number formats
// Examples: "Op. 67", "BWV 1080", "K. 550", "Hob. XVI:52"
var opusPattern = regexp.MustCompile(`(?i)\b(Op\.?|BWV|K\.?|Hob\.?|D\.?|RV|Wq\.?|S\.?)\s*\d+`)

// OpusNumbers checks for presence of opus/catalog numbers (classical.opus)
// INFO level - suggests including catalog numbers for better identification
func (r *Rules) OpusNumbers(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.opus",
		Name:   "Opus/catalog numbers recommended in track titles",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

	var issues []domain.ValidationIssue

	// Check if reference has opus numbers but actual doesn't
	if reference != nil {
		refTracks := reference.Tracks
		actualTracks := actual.Tracks

		refTrackMap := make(map[string]*domain.Track)
		for _, rt := range refTracks {
			key := fmt.Sprintf("%d-%d", rt.Disc, rt.Track)
			refTrackMap[key] = rt
		}

		for _, actualTrack := range actualTracks {
			key := fmt.Sprintf("%d-%d", actualTrack.Disc, actualTrack.Track)
			refTrack, exists := refTrackMap[key]

			if !exists {
				continue
			}

			// Check if reference has opus but actual doesn't
			refHasOpus := hasOpusNumber(refTrack.Title)
			actualHasOpus := hasOpusNumber(actualTrack.Title)

			if refHasOpus && !actualHasOpus {
				opusNum := extractOpusNumber(refTrack.Title)
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelInfo,
					Track: actualTrack.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s: Consider adding opus/catalog number '%s' from reference",
						formatTrackNumber(actualTrack), opusNum),
				})
			}
		}
	} else {
		// No reference - just suggest adding opus numbers if missing
		for _, track := range actual.Tracks {
			if !hasOpusNumber(track.Title) {
				// Only suggest for composers known to have catalog systems
				composer := getComposer(track.Artists)
				if needsCatalogNumber(composer) {
					issues = append(issues, domain.ValidationIssue{
						Level: domain.LevelInfo,
						Track: track.Track,
						Rule:  meta.ID,
						Message: fmt.Sprintf("Track %s: Consider adding opus/catalog number for better identification",
							formatTrackNumber(track)),
					})
				}
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// hasOpusNumber checks if title contains opus/catalog number
func hasOpusNumber(title string) bool {
	return opusPattern.MatchString(title)
}

// extractOpusNumber extracts the opus/catalog number from title
func extractOpusNumber(title string) string {
	match := opusPattern.FindString(title)
	return strings.TrimSpace(match)
}

// needsCatalogNumber checks if composer typically has catalog numbers
func needsCatalogNumber(composer string) bool {
	if composer == "" {
		return false
	}

	lowerComposer := strings.ToLower(composer)

	// Composers with well-known catalog systems
	catalogComposers := []string{
		"beethoven", "mozart", "bach", "schubert", "haydn",
		"vivaldi", "handel", "telemann", "brahms", "chopin",
		"liszt", "schumann", "mendelssohn", "dvorak",
	}

	for _, cat := range catalogComposers {
		if strings.Contains(lowerComposer, cat) {
			return true
		}
	}

	return false
}
