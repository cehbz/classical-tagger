package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_RequiredTags(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		actual       *domain.Album
		wantPass     bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name: "valid - all required tags present",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Ludwig van Beethoven", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Symphony No. 5", []domain.Artist{composer, ensemble})
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass: true,
		},
		{
			name: "missing year - warning only",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})
				album, _ := domain.NewAlbum("Beethoven Symphonies", 0)
				album.AddTrack(track)
				return album
			}(),
			wantPass:     false,
			wantErrors:   0,
			wantWarnings: 1,
		},
		{
			name: "missing track title",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "", []domain.Artist{composer, ensemble})
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "missing artists",
			actual: func() *domain.Album {
				track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{})
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "only composer, no performers",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
				track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer})
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "multiple tracks, some missing title",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)
				artists := []domain.Artist{composer, ensemble}

				track1, _ := domain.NewTrack(1, 1, "Symphony No. 1", artists)
				track2, _ := domain.NewTrack(1, 2, "", artists)
				track3, _ := domain.NewTrack(1, 3, "Symphony No. 3", artists)

				album, _ := domain.NewAlbum("Beethoven", 1963)
				album.AddTrack(track1)
				album.AddTrack(track2)
				album.AddTrack(track3)
				return album
			}(),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "multiple issues",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)
				artists := []domain.Artist{composer, ensemble}

				track1, _ := domain.NewTrack(1, 1, "", artists)
				track2, _ := domain.NewTrack(1, 2, "Symphony", []domain.Artist{})

				album, _ := domain.NewAlbum("", 0)
				album.AddTrack(track1)
				album.AddTrack(track2)
				return album
			}(),
			wantPass:     false,
			wantErrors:   3, // Album title, track1 title, track2 artists
			wantWarnings: 1, // Year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.RequiredTags(tt.actual, tt.actual)

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
				if warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}
