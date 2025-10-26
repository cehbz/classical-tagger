package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArrangerCredit(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name     string
		actual   *domain.Album
		wantPass bool
		wantInfo int
	}{
		{
			name: "valid - no arranger",
			actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Orchestra", domain.RoleEnsemble,
			),
			wantPass: true,
		},
		{
			name:     "info - arranger present, no credit in title",
			actual:   buildAlbumWithTitleAndArranger("Symphony No. 5", "Mahler"),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "valid - arrangement credited in title",
			actual:   buildAlbumWithTitleAndArranger("Symphony No. 5, arr. Mahler", "Gustav Mahler"),
			wantPass: true,
		},
		{
			name:     "valid - arranged by in title",
			actual:   buildAlbumWithTitleAndArranger("Prelude, arranged by Busoni", "Ferruccio Busoni"),
			wantPass: true,
		},
		{
			name:     "valid - transcription credited",
			actual:   buildAlbumWithTitleAndArranger("Chaconne, transcription by Brahms", "Johannes Brahms"),
			wantPass: true,
		},
		{
			name:     "info - has arr. but wrong arranger name",
			actual:   buildAlbumWithTitleAndArranger("Prelude, arr. Bach", "Ferruccio Busoni"),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "valid - last name sufficient",
			actual:   buildAlbumWithTitleAndArranger("Theme, arr. Rachmaninoff", "Sergei Rachmaninoff"),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ArrangerCredit(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				infoCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelInfo {
						infoCount++
					}
				}

				if infoCount != tt.wantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.wantInfo)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

// buildAlbumWithTitleAndArranger creates album with specific track title and arranger
func buildAlbumWithTitleAndArranger(trackTitle, arrangerName string) *domain.Album {
	composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
	arranger, _ := domain.NewArtist(arrangerName, domain.RoleArranger)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)

	artists := []domain.Artist{composer, arranger, ensemble}
	track, _ := domain.NewTrack(1, 1, trackTitle, artists)
	album, _ := domain.NewAlbum("Album", 1963)
	album.AddTrack(track)
	return album
}
