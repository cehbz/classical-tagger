package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_YearFieldUsage(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		actual       *domain.Album
		reference    *domain.Album
		wantPass     bool
		wantErrors   int
		wantWarnings int
		wantInfo     int
	}{
		{
			name:     "valid - reasonable year",
			actual:   buildAlbumWithTitle("Symphony", "1963"),
			wantPass: true,
		},
		{
			name:         "warning - year too early",
			actual:       buildAlbumWithTitle("Symphony", "1850"),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "error - year in future",
			actual:     buildAlbumWithTitle("Symphony", "2030"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:      "valid - matches reference",
			actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			reference: buildAlbumWithTitle("Symphony", "No. 9"),
			wantPass:  true,
		},
		{
			name:      "info - small difference from reference",
			actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			reference: buildAlbumWithTitle("Symphony", "No 9"),
			wantPass:  false,
			wantInfo:  1,
		},
		{
			name:         "warning - large difference from reference",
			actual:       buildAlbumWithTitle("Symphony", "No. 9"),
			reference:    buildAlbumWithTitle("Symphony", "1980"),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - edition year before album year",
			actual:       buildAlbumWithEditionYear(1990, 1985),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:     "valid - edition year after album year",
			actual:   buildAlbumWithEditionYear(1963, 2010),
			wantPass: true,
		},
		{
			name:     "valid - same edition and album year",
			actual:   buildAlbumWithEditionYear(1963, 1963),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.YearFieldUsage(tt.actual, tt.reference)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				errorCount := 0
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues() {
					switch issue.Level() {
					case domain.LevelError:
						errorCount++
					case domain.LevelWarning:
						warningCount++
					case domain.LevelInfo:
						infoCount++
					}
				}

				if tt.wantErrors > 0 && errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}
				if tt.wantWarnings > 0 && warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}
				if tt.wantInfo > 0 && infoCount != tt.wantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.wantInfo)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

// buildAlbumWithEditionYear creates album with specific album and edition years
func buildAlbumWithEditionYear(albumYear, editionYear int) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})

	edition, _ := domain.NewEdition("Label", editionYear)
	edition = edition.WithCatalogNumber("CAT123")
	album, _ := domain.NewAlbum("Album", albumYear)
	album = album.WithEdition(edition)
	album.AddTrack(track)
	return album
}
