package validation

import (
	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumHasTracks checks that the album contains at least one track (album: 2.3.16.4-album-tracks)
// This is an ALBUM-LEVEL rule - signature: (actual, reference *Album)
func (r *Rules) AlbumHasTracks(actual, _ *domain.Album) RuleResult {
    meta := RuleMetadata{
        ID:     "2.3.16.4",
		Name:   "Album must have at least one track",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	if len(actual.Tracks) == 0 {
		issue := domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: "Album must have at least one track",
		}
		return RuleResult{Meta: meta, Issues: []domain.ValidationIssue{issue}}
	}

	return RuleResult{Meta: meta, Issues: nil}
}