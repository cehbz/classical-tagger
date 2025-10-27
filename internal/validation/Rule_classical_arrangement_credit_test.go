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
			Name: "valid - no arranger",
			Actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Orchestra", domain.RoleEnsemble,
			),
			WantPass: true,
		},
		{
			Name:     "info - arranger present, no credit in title",
			Actual:   buildAlbumWithTitleAndArranger("Symphony No. 5", "Mahler"),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - arrangement credited in title",
			Actual:   buildAlbumWithTitleAndArranger("Symphony No. 5, arr. Mahler", "Gustav Mahler"),
			WantPass: true,
		},
		{
			Name:     "valid - arranged by in title",
			Actual:   buildAlbumWithTitleAndArranger("Prelude, arranged by Busoni", "Ferruccio Busoni"),
			WantPass: true,
		},
		{
			Name:     "valid - transcription credited",
			Actual:   buildAlbumWithTitleAndArranger("Chaconne, transcription by Brahms", "Johannes Brahms"),
			WantPass: true,
		},
		{
			Name:     "info - has arr. but wrong arranger name",
			Actual:   buildAlbumWithTitleAndArranger("Prelude, arr. Bach", "Ferruccio Busoni"),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - last name sufficient",
			Actual:   buildAlbumWithTitleAndArranger("Theme, arr. Rachmaninoff", "Sergei Rachmaninoff"),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.ArrangerCredit(tt.Actual, tt.Actual)

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

// buildAlbumWithTitleAndArranger creates album with specific track title and arranger
func buildAlbumWithTitleAndArranger(trackTitle, arrangerName string) *domain.Album {
	return &domain.Album{
		Title: "Album", 
		OriginalYear: 1963, 
		Tracks: []*domain.Track{
			{
				Disc: 1, 
				Track: 1, 
				Title: trackTitle, 
				Artists: []domain.Artist{
					{Name: "Bach", Role: domain.RoleComposer}, 
					{Name: arrangerName, Role: domain.RoleArranger}, 
					{Name: "Orchestra", Role: domain.RoleEnsemble}},
			},
		},
	}
}
