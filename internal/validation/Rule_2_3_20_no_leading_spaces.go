package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// NoLeadingSpaces checks that no file or folder names have leading spaces (rule 2.3.20)
func (r *Rules) NoLeadingSpaces(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.20",
		Name:   "No leading spaces in file or folder names",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check album title (represents folder name)
	if strings.HasPrefix(actual.Title, " ") {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0, // Album-level issue
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title has leading space: '%s'", actual.Title),
		})
	}

	// Check each track filename and title
	for _, track := range actual.Tracks {
		// Check filename
		fileName := track.Name
		if fileName != "" {
			// Check the base filename and any path components
			parts := strings.Split(fileName, "/")
			for i, part := range parts {
				if strings.HasPrefix(part, " ") {
					var location string
					if i == len(parts)-1 {
						location = "filename"
					} else {
						location = "folder name"
					}
					issues = append(issues, domain.ValidationIssue{
						Level: domain.LevelError,
						Track: track.Track,
						Rule:  meta.ID,
						Message: fmt.Sprintf("Track %s has leading space in %s: '%s'",
							formatTrackNumber(track), location, part),
					})
				}
			}
		}

		// Check track title (tag value)
		if strings.HasPrefix(track.Title, " ") {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s title tag has leading space: '%s'",
					formatTrackNumber(track), track.Title),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// formatTrackNumber formats a track for error messages
func formatTrackNumber(track *domain.Track) string {
	if track.Disc > 1 {
		return fmt.Sprintf("%d-%d", track.Disc, track.Track)
	}
	return fmt.Sprintf("%d", track.Track)
}
