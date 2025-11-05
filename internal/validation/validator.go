package validation

import "github.com/cehbz/classical-tagger/internal/domain"

// Check validates a torrent's metadata against validation rules.
// If reference is nil, only non-reference-dependent validations are performed.
// Returns all validation issues found.
func Check(actual, reference *domain.Torrent) []domain.ValidationIssue {
	var issues []domain.ValidationIssue
	rules := NewRules()

	// Run all torrent-level rules
	torrentRules := rules.TorrentRules()
	for _, rule := range torrentRules {
		result := rule(actual, reference)
		issues = append(issues, result.Issues...)
	}

	// Run all track-level rules
	trackRules := rules.TrackRules()

	// Iterate through tracks and validate each one
	actualTracks := actual.Tracks()
	refTracks := []*domain.Track(nil)
	if reference != nil {
		refTracks = reference.Tracks()
	}

	for i, actualTrack := range actualTracks {
		var refTrack *domain.Track
		if i < len(refTracks) {
			refTrack = refTracks[i]
		}

		// Run each track rule for this track
		for _, rule := range trackRules {
			result := rule(actualTrack, refTrack, actual, reference)
			issues = append(issues, result.Issues...)
		}
	}

	return issues
}
