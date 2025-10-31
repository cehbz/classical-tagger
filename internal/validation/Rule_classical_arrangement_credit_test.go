package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArrangerCredit(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Album
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "valid - no arranger",
			Actual:   NewAlbum().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "info - arranger present, no credit in title",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Mahler", Role: domain.RoleArranger}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - arrangement credited in title",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5, arr. Mahler").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Gustav Mahler", Role: domain.RoleArranger}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - arranged by in title",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Prelude, arranged by Busoni").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Ferruccio Busoni", Role: domain.RoleArranger}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - transcription credited",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Chaconne, transcription by Brahms").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Johannes Brahms", Role: domain.RoleArranger}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "info - has arr. but wrong arranger name",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Prelude, arr. Bach").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Ferruccio Busoni", Role: domain.RoleArranger}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - last name sufficient",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Theme, arr. Rachmaninoff").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Rachmaninoff", Role: domain.RoleArranger}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		for _, track := range tt.Actual.Tracks {
			t.Run(tt.Name, func(t *testing.T) {
				result := rules.ArrangerCredit(track, nil, nil, nil)

				if result.Passed() != tt.WantPass {
					t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
				}

				if !tt.WantPass {
					infoCount := 0
					for _, issue := range result.Issues {
						if issue.Level == domain.LevelInfo {
							infoCount++
						}
					}

					if infoCount != tt.WantInfo {
						t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
					}

					for _, issue := range result.Issues {
						t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
					}
				}
			})
		}
	}
}
