package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// opusPattern matches opus number formats
// Examples: "Op. 67", "BWV 1080", "K. 550", "Hob. XVI:52"
// Accept Hoboken Roman:Arabic form like "Hob. XVI:52"
var opusPattern = regexp.MustCompile(`(?i)\b(Op\.?|BWV|K\.?|Hob\.?|D\.?|RV|Wq\.?|S\.?)\s*([IVXLCDM]+:)?\s*\d+`)

// OpusNumbers checks for presence of opus/catalog numbers (classical.opus)
// INFO level - suggests including catalog numbers for better identification
func (r *Rules) OpusNumbers(actualTrack, refTrack *domain.Track, _, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.opus",
		Name:   "Opus/catalog numbers recommended in track titles",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

	if refTrack == nil {
		// No reference - just suggest adding opus numbers if missing

		if extractOpusNumber(actualTrack.Title) != "" {
			return RuleResult{Meta: meta, Issues: nil}
		}
		// Only suggest for composers known to have catalog systems
		composer := getComposer(actualTrack.Artists)
		if !needsCatalogNumber(composer) {
			return RuleResult{Meta: meta, Issues: nil}
		}
		return RuleResult{Meta: meta, Issues: []domain.ValidationIssue{{
			Level: domain.LevelInfo,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Consider adding opus/catalog number for better identification",
				formatTrackNumber(actualTrack)),
		}}}
	}
	// Check if reference has opus but actual doesn't
	reOpus := extractOpusNumber(refTrack.Title)
	actualOpus := extractOpusNumber(actualTrack.Title)

	if reOpus == actualOpus {
		return RuleResult{Meta: meta, Issues: nil}
	}

	return RuleResult{Meta: meta, Issues: []domain.ValidationIssue{{
		Level: domain.LevelInfo,
		Track: actualTrack.Track,
		Rule:  meta.ID,
		Message: fmt.Sprintf("Track %s: Opus %s doesn't match reference %s",
			formatTrackNumber(actualTrack), actualOpus, reOpus),
	}}}
}

// extractOpusNumber extracts the opus/catalog number from title
func extractOpusNumber(title string) string {
	match := opusPattern.FindStringSubmatch(title)
	if len(match) == 0 {
		return ""
	}
	full := match[0]
	return strings.TrimSpace(full)
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
