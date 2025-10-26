package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoLeadingSpaces(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - no leading spaces",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
				track1, _ := domain.NewTrack(1, 1, "Symphony No. 1", []domain.Artist{composer, ensemble})
				track1 = track1.WithName("01 - Symphony.flac")
				track2, _ := domain.NewTrack(1, 2, "Symphony No. 2", []domain.Artist{composer, ensemble})
				track2 = track2.WithName("02 - Concerto.flac")
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track1)
				album.AddTrack(track2)
				return album
			}(),
			wantPass: true,
		},
		{
			name: "album title with leading space",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})
				track = track.WithName("01 - Symphony.flac")
				album, _ := domain.NewAlbum(" Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "filename with leading space",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})
				track = track.WithName(" 01 - Symphony.flac")
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "track title with leading space",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, " Symphony No. 1", []domain.Artist{composer, ensemble})
				track = track.WithName("01 - Symphony.flac")
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "folder name in path with leading space",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})
				track = track.WithName(" CD1/01 - Symphony.flac")
				album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
				album.AddTrack(track)
				return album
			}(),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "multiple violations",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
				track1, _ := domain.NewTrack(1, 1, " Symphony No. 1", []domain.Artist{composer, ensemble})
				track1 = track1.WithName(" 01 - Symphony.flac")
				track2, _ := domain.NewTrack(1, 2, "Concerto", []domain.Artist{composer, ensemble})
				track2 = track2.WithName("02 - Concerto.flac")
				album, _ := domain.NewAlbum(" Beethoven", 1963)
				album.AddTrack(track1)
				album.AddTrack(track2)
				return album
			}(),
			wantPass:   false,
			wantIssues: 3, // Album title + track1 filename + track1 title
		},
		{
			name: "valid multi-disc with subfolders",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
				track1, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})
				track1 = track1.WithName("CD1/01 - Symphony.flac")
				track2, _ := domain.NewTrack(2, 1, "Concerto", []domain.Artist{composer, ensemble})
				track2 = track2.WithName("CD2/01 - Concerto.flac")
				album, _ := domain.NewAlbum("Beethoven", 1963)
				album.AddTrack(track1)
				album.AddTrack(track2)
				return album
			}(),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.NoLeadingSpaces(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass && len(result.Issues()) != tt.wantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues()), tt.wantIssues)
				for _, issue := range result.Issues() {
					t.Logf("  Issue: %s", issue.Message())
				}
			}
		})
	}
}
