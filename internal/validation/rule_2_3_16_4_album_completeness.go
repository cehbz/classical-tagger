package validation

import (
	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumHasTracks checks that the album contains at least one track (rule 2.3.16.4)
func (r *Rules) AlbumHasTracks(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.16.4",
		name:   "Album must have at least one track",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	if len(actual.Tracks()) == 0 {
		issue := domain.NewIssue(
			domain.LevelError,
			0, // album-level
			meta.id,
			"Album must have at least one track",
		)
		return meta.Fail(issue)
	}
	
	return meta.Pass()
}