package domain

// Album represents a classical music release.
// All fields are exported and mutable.
type Album struct {
	Title        string    `json:"title"`
	OriginalYear int       `json:"original_year"`
	Edition      *Edition  `json:"edition,omitempty"`
	Tracks       []*Track  `json:"tracks"`
}

// Validate checks the entire album for compliance with rules.
// Returns all validation issues from the album and all its tracks.
func (a *Album) Validate() []ValidationIssue {
	var issues []ValidationIssue

	// Album must have at least one track (2.3.16.4 implies music content)
	if len(a.Tracks) == 0 {
		issues = append(issues, ValidationIssue{
			Level:   LevelError,
			Track:   0, // album-level
			Rule:    "2.3.16.4",
			Message: "Album must have at least one track",
		})
	}

	// Edition is optional but strongly recommended (Classical Guide preamble)
	if a.Edition == nil {
		issues = append(issues, ValidationIssue{
			Level:   LevelWarning,
			Track:   0, // album-level
			Rule:    "Classical Guide: Step 3",
			Message: "Edition information (label, catalog number) is strongly recommended",
		})
	} else {
		// Validate edition if present
		editionIssues := a.Edition.Validate()
		issues = append(issues, editionIssues...)
	}

	// Year tag is optional but strongly encouraged (2.3.16.4).
	// Warn if unknown, basic sanity otherwise can be checked elsewhere if needed.
	if a.OriginalYear == 0 {
		issues = append(issues, ValidationIssue{
			Level:   LevelWarning,
			Track:   0,
			Rule:    "2.3.16.4",
			Message: "Year is optional but strongly encouraged; consider adding original release year",
		})
	}

	// Validate all tracks
	for _, track := range a.Tracks {
		trackIssues := track.Validate()
		issues = append(issues, trackIssues...)
	}

	return issues
}
