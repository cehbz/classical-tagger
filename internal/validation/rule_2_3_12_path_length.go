package validation

import (
	"fmt"
	"path/filepath"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// PathLength checks that file paths don't exceed 180 characters (rule 2.3.12)
func (r *Rules) PathLength(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.12",
		Name:   "Path length under 180 characters",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check each track's path length
	// For now we don't have folderName() in Album, so we'll use Title() as a proxy
	baseFolder := actual.Title

	for _, track := range actual.Tracks {
		fileName := track.Name
		if fileName == "" {
			continue // Skip tracks without filenames
		}

		path := filepath.Join(baseFolder, fileName)
		if len(path) > 180 {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   track.Track,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Path length %d exceeds 180 characters: %s", len(path), path),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
