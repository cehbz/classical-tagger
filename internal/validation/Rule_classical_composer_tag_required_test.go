package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerTagRequired(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		WantPass     bool
		WantErrors   int
		WantWarnings int
	}{
		{
			Name:     "valid - full composer name",
			Actual:   buildAlbumWithComposer("Ludwig van Beethoven"),
			WantPass: true,
		},
		{
			Name:     "valid - composer with initials",
			Actual:   buildAlbumWithComposer("J.S. Bach"),
			WantPass: true,
		},
		{
			Name:     "valid - composer with spaced initials",
			Actual:   buildAlbumWithComposer("J. S. Bach"),
			WantPass: true,
		},
		{
			Name:     "valid - two-word name",
			Actual:   buildAlbumWithComposer("Johann Bach"),
			WantPass: true,
		},
		{
			Name:     "valid - composer with surname prefix",
			Actual:   buildAlbumWithComposer("Wolfgang Amadeus Mozart"),
			WantPass: true,
		},
		{
			Name:       "invalid - last name only (ambiguous)",
			Actual:     buildAlbumWithComposer("Bach"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "invalid - missing composer",
			Actual: func() *domain.Album {
				ensemble := domain.Artist{Name: "Vienna Phil", Role: domain.RoleEnsemble}
				track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{ensemble}}
				return &domain.Album{Title: "Symphonies", OriginalYear: 1963, Tracks: []*domain.Track{&track}}
			}(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "multiple tracks, one missing composer",
			Actual: func() *domain.Album {
				composer := domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}
				ensemble := domain.Artist{Name: "Vienna Phil", Role: domain.RoleEnsemble}

				track1 := domain.Track{Disc: 1, Track: 1, Title: "Symphony No. 1", Artists: []domain.Artist{composer, ensemble}}
				track2 := domain.Track{Disc: 1, Track: 2, Title: "Symphony No. 2", Artists: []domain.Artist{ensemble}}
				track3 := domain.Track{Disc: 1, Track: 3, Title: "Symphony No. 3", Artists: []domain.Artist{composer, ensemble}}

				return &domain.Album{Title: "Beethoven Symphonies", OriginalYear: 1963, Tracks: []*domain.Track{&track1, &track2, &track3}}
			}(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "multiple tracks, some ambiguous names",
			Actual: func() *domain.Album {
				composer1 := domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer}
				composer2 := domain.Artist{Name: "Bach", Role: domain.RoleComposer} // Ambiguous
				ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

				track1 := domain.Track{Disc: 1, Track: 1, Title: "Work 1", Artists: []domain.Artist{composer1, ensemble}}
				track2 := domain.Track{Disc: 1, Track: 2, Title: "Work 2", Artists: []domain.Artist{composer2, ensemble}}

				return &domain.Album{Title: "Bach Works", OriginalYear: 1963, Tracks: []*domain.Track{&track1, &track2}}
			}(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "edge case - Beethoven, Ludwig van (reversed format)",
			Actual:   buildAlbumWithComposer("Beethoven, Ludwig van"),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.ComposerTagRequired(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelError {
						errorCount++
					} else if issue.Level == domain.LevelWarning {
						warningCount++
					}
				}

				if errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}
				if tt.WantWarnings > 0 && warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

// Helper function to build an album with a specific composer name
func buildAlbumWithComposer(composerName string) *domain.Album {
	composer := domain.Artist{Name: composerName, Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}
	conductor := domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}

	artists := []domain.Artist{composer, ensemble, conductor}
	track := domain.Track{Disc: 1, Track: 1, Title: "Symphony No. 5", Artists: artists}
	return &domain.Album{Title: "Beethoven Symphonies", OriginalYear: 1963, Tracks: []*domain.Track{&track}}
}
