package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerTagRequired(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		actual       *domain.Album
		wantPass     bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name:     "valid - full composer name",
			actual:   buildAlbumWithComposer("Ludwig van Beethoven"),
			wantPass: true,
		},
		{
			name:     "valid - composer with initials",
			actual:   buildAlbumWithComposer("J.S. Bach"),
			wantPass: true,
		},
		{
			name:     "valid - composer with spaced initials",
			actual:   buildAlbumWithComposer("J. S. Bach"),
			wantPass: true,
		},
		{
			name:     "valid - two-word name",
			actual:   buildAlbumWithComposer("Johann Bach"),
			wantPass: true,
		},
		{
			name:     "valid - composer with surname prefix",
			actual:   buildAlbumWithComposer("Wolfgang Amadeus Mozart"),
			wantPass: true,
		},
		{
			name:       "invalid - last name only (ambiguous)",
			actual:     buildAlbumWithComposer("Bach"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "invalid - missing composer",
			actual: func() *domain.Album {
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{ensemble})
				album, _ := domain.NewAlbum("Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "multiple tracks, one missing composer",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Ludwig van Beethoven", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)

				track1, _ := domain.NewTrack(1, 1, "Symphony No. 1", []domain.Artist{composer, ensemble})
				track2, _ := domain.NewTrack(1, 2, "Symphony No. 2", []domain.Artist{ensemble})
				track3, _ := domain.NewTrack(1, 3, "Symphony No. 3", []domain.Artist{composer, ensemble})

				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track1)
				album.AddTrack(track2)
				album.AddTrack(track3)
				return album
			}(),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "multiple tracks, some ambiguous names",
			actual: func() *domain.Album {
				composer1, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
				composer2, _ := domain.NewArtist("Bach", domain.RoleComposer) // Ambiguous
				ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)

				track1, _ := domain.NewTrack(1, 1, "Work 1", []domain.Artist{composer1, ensemble})
				track2, _ := domain.NewTrack(1, 2, "Work 2", []domain.Artist{composer2, ensemble})

				album, _ := domain.NewAlbum("Bach Works", 1963)
				album.AddTrack(track1)
				album.AddTrack(track2)
				return album
			}(),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:     "edge case - Beethoven, Ludwig van (reversed format)",
			actual:   buildAlbumWithComposer("Beethoven, Ludwig van"),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ComposerTagRequired(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelError {
						errorCount++
					} else if issue.Level() == domain.LevelWarning {
						warningCount++
					}
				}

				if errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}
				if tt.wantWarnings > 0 && warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

// Helper function to build an album with a specific composer name
func buildAlbumWithComposer(composerName string) *domain.Album {
	composer, _ := domain.NewArtist(composerName, domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Vienna Philharmonic", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Herbert von Karajan", domain.RoleConductor)

	artists := []domain.Artist{composer, ensemble, conductor}
	track, _ := domain.NewTrack(1, 1, "Symphony No. 5", artists)
	album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
	album.AddTrack(track)
	return album
}
