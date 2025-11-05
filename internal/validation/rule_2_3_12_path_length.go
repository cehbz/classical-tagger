package validation

import (
	"fmt"
	"path/filepath"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// PathLength checks that file paths don't exceed 180 characters (rule 2.3.12)
func (r *Rules) PathLength(actualTrack, _ *domain.Track, actualTorrent, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.12",
		Name:   "Path length under 180 characters",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	path := filepath.Join(actualTorrent.RootPath, actualTrack.File.Path)
	if len(path) > 180 {
		return RuleResult{Meta: meta, Issues: []domain.ValidationIssue{
			{
				Level:   domain.LevelError,
				Track:   actualTrack.Track,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Path length %d exceeds 180 characters: %s", len(path), path),
			},
		}}
	}
	return RuleResult{Meta: meta, Issues: nil}
}
