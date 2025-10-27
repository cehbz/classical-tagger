package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_YearFieldUsage(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		Reference    *domain.Album
		WantPass     bool
		WantErrors   int
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:     "valid - reasonable year",
			Actual:   buildAlbumWithTitle("Symphony", "1963"),
			WantPass: true,
		},
		{
			Name:         "warning - year too early",
			Actual:       buildAlbumWithTitle("Symphony", "1850"),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "error - year in future",
			Actual:     buildAlbumWithTitle("Symphony", "2030"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:      "valid - matches reference",
			Actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			Reference: buildAlbumWithTitle("Symphony", "No. 9"),
			WantPass:  true,
		},
		{
			Name:      "info - small difference from reference",
			Actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			Reference: buildAlbumWithTitle("Symphony", "No 9"),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:         "warning - large difference from reference",
			Actual:       buildAlbumWithTitle("Symphony", "No. 9"),
			Reference:    buildAlbumWithTitle("Symphony", "1980"),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - edition year before album year",
			Actual:       buildAlbumWithEditionYear(1990, 1985),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:     "valid - edition year after album year",
			Actual:   buildAlbumWithEditionYear(1963, 2010),
			WantPass: true,
		},
		{
			Name:     "valid - same edition and album year",
			Actual:   buildAlbumWithEditionYear(1963, 1963),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.YearFieldUsage(tt.Actual, tt.Reference)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues {
					switch issue.Level {
					case domain.LevelError:
						errorCount++
					case domain.LevelWarning:
						warningCount++
					case domain.LevelInfo:
						infoCount++
					}
				}

				if tt.WantErrors > 0 && errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}
				if tt.WantWarnings > 0 && warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}
				if tt.WantInfo > 0 && infoCount != tt.WantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

// buildAlbumWithEditionYear creates album with specific album and edition years
func buildAlbumWithEditionYear(albumYear, editionYear int) *domain.Album {
	return &domain.Album{
		Title:        "Album",
		OriginalYear: albumYear,
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Symphony",
				Artists: []domain.Artist{
					domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
					domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
				},
			},
		},
		Edition: &domain.Edition{
			Label:         "Label",
			Year:          editionYear,
			CatalogNumber: "CAT123",
		},
	}
}
