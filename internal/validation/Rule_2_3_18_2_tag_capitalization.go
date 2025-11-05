package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumTagCapitalization checks that album tags use proper Title Case (album: 2.3.18.2-album)
func (r *Rules) AlbumTagCapitalization(actualTorrent, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.2-album",
		Name:   "Tags must use proper Title Case capitalization",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check album title
	if capIssue := checkCapitalization(actualTorrent.Title); capIssue != "" {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title '%s': %s", actualTorrent.Title, capIssue),
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// TrackTagCapitalization checks that track tags use proper Title Case (track: 2.3.18.2)
func (r *Rules) TrackTagCapitalization(actualTrack, _ *domain.Track, _, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.2",
		Name:   "Tags must use proper Title Case capitalization",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check track title
	if capIssue := checkCapitalization(actualTrack.Title); capIssue != "" {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   actualTrack.Track,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Track %s title '%s': %s", formatTrackNumber(actualTrack), actualTrack.Title, capIssue),
		})
	}

	// Check artist names
	for _, artist := range actualTrack.Artists {
		if capIssue := checkCapitalization(artist.Name); capIssue != "" {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s artist '%s': %s",
					formatTrackNumber(actualTrack), artist.Name, capIssue),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// TagCapitalizationVsReference checks capitalization matches reference (rule 2.3.18.2 with reference)
func (r *Rules) TagCapitalizationVsReference(actual, reference *domain.Torrent) RuleResult {
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
	actualTracks := actual.Tracks()
	refTracks := reference.Tracks()

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
