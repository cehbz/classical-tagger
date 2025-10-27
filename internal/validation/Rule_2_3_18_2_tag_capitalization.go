package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TagCapitalization checks that tags use proper Title Case (rule 2.3.18.2)
func (r *Rules) TagCapitalization(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.2",
		Name:   "Tags must use proper Title Case capitalization",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check album title
	if capIssue := checkCapitalization(actual.Title); capIssue != "" {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title: %s", capIssue),
		})
	}

	// Check each track
	for _, track := range actual.Tracks {
		// Check track title
		if capIssue := checkCapitalization(track.Title); capIssue != "" {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   track.Track,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Track %s title: %s", formatTrackNumber(track), capIssue),
			})
		}

		// Check artist names
		for _, artist := range track.Artists {
			if capIssue := checkCapitalization(artist.Name); capIssue != "" {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelError,
					Track: track.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s artist '%s': %s",
						formatTrackNumber(track), artist.Name, capIssue),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// TagCapitalizationVsReference checks capitalization matches reference (rule 2.3.18.2 with reference)
func (r *Rules) TagCapitalizationVsReference(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.2.ref",
		Name:   "Tag capitalization should match reference",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	if reference == nil {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	// Compare album title capitalization
	if actual.Title != "" && reference.Title != "" {
		if !capitalizationMatches(actual.Title, reference.Title) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelWarning,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Album title capitalization differs from reference: '%s' vs '%s'",
					actual.Title, reference.Title),
			})
		}
	}

	// Compare track titles
	actualTracks := actual.Tracks
	refTracks := reference.Tracks

	refTrackMap := make(map[string]*domain.Track)
	for _, refTrack := range refTracks {
		key := fmt.Sprintf("%d-%d", refTrack.Disc, refTrack.Track)
		refTrackMap[key] = refTrack
	}

	for _, actualTrack := range actualTracks {
		key := fmt.Sprintf("%d-%d", actualTrack.Disc, actualTrack.Track)
		refTrack, exists := refTrackMap[key]

		if !exists {
			continue
		}

		if !capitalizationMatches(actualTrack.Title, refTrack.Title) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelWarning,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Capitalization differs from reference: '%s' vs '%s'",
					formatTrackNumber(actualTrack), actualTrack.Title, refTrack.Title),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// capitalizationMatches checks if two strings have the same capitalization pattern
func capitalizationMatches(s1, s2 string) bool {
	// If normalized versions match, capitalization is compatible
	if normalizeTitle(s1) == normalizeTitle(s2) {
		// But check if actual case differs
		return s1 == s2
	}
	return false
}
