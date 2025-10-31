package validation

import "github.com/cehbz/classical-tagger/internal/domain"

// Check validates an album's metadata against validation rules.
// If reference is nil, only non-reference-dependent validations are performed.
// Returns all validation issues found.
func Check(actual, reference *domain.Album) []domain.ValidationIssue {
	var issues []domain.ValidationIssue
	rules := NewRules()

	// Run all album-level rules
	albumRules := rules.AlbumRules()
	for _, rule := range albumRules {
		result := rule(actual, reference)
		issues = append(issues, result.Issues...)
	}

	// Run all track-level rules
	trackRules := rules.TrackRules()

	// Iterate through tracks and validate each one
	for i, actualTrack := range actual.Tracks {
		var refTrack *domain.Track
		if reference != nil && i < len(reference.Tracks) {
			refTrack = reference.Tracks[i]
		}

		// Run each track rule for this track
		for _, rule := range trackRules {
			result := rule(actualTrack, refTrack, actual, reference)
			issues = append(issues, result.Issues...)
		}
	}

	return issues
}
