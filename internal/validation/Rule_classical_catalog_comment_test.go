package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CatalogInfoInComments(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name     string
		actual   *domain.Album
		wantPass bool
		wantInfo int
	}{
		{
			name:     "pass - complete edition info",
			actual:   buildAlbumWithCompleteEdition(),
			wantPass: true,
		},
		{
			name:     "info - no edition info",
			actual:   buildAlbumWithTitle("Symphony", "1963"),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - missing label",
			actual:   buildAlbumWithPartialEdition("", "CAT123", 1990),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - missing catalog",
			actual:   buildAlbumWithPartialEdition("Deutsche Grammophon", "", 1990),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - missing year",
			actual:   buildAlbumWithPartialEdition("Deutsche Grammophon", "CAT123", 0),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - multiple missing fields",
			actual:   buildAlbumWithPartialEdition("", "", 0),
			wantPass: false,
			wantInfo: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CatalogInfoInComments(tt.actual, tt.actual)

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

// buildAlbumWithCompleteEdition creates album with complete edition information
func buildAlbumWithCompleteEdition() *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})

	edition, _ := domain.NewEdition("Deutsche Grammophon", 1990)
	edition = edition.WithCatalogNumber("DG-479-0334")
	album, _ := domain.NewAlbum("Album", 1963)
	album.AddTrack(track)
	album.WithEdition(edition)
	return album
}

// buildAlbumWithPartialEdition creates album with partial edition information
func buildAlbumWithPartialEdition(label, catalog string, year int) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})

	edition, _ := domain.NewEdition(label, year)
	edition = edition.WithCatalogNumber(catalog)
	album, _ := domain.NewAlbum("Album", 1963)
	album.AddTrack(track)
	album.WithEdition(edition)
	return album
}
