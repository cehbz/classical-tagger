package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TagAccuracyVsReference(t *testing.T) {
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
			Name:      "valid - exact match",
			Actual:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			Reference: buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			WantPass:  true,
		},
		{
			Name:         "warning - year mismatch",
			Actual:       buildBasicAlbum("Symphony No. 5", 1960, "Beethoven"),
			Reference:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "error - composer mismatch",
			Actual:     buildBasicAlbum("Symphony No. 5", 1963, "Mozart"),
			Reference:  buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - title mismatch",
			Actual:     buildBasicAlbum("Symphony No. 6", 1963, "Beethoven"),
			Reference:  buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:      "info - minor title difference",
			Actual:    buildBasicAlbum("Sympony No. 5", 1963, "Beethoven"), // Typo
			Reference: buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:         "warning - moderate title difference",
			Actual:       buildBasicAlbum("Symphony No. 5 Finale", 1963, "Beethoven"),
			Reference:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:      "pass - no reference",
			Actual:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			Reference: nil,
			WantPass:  true,
		},
		{
			Name:         "multiple errors",
			Actual:       buildBasicAlbum("Concerto", 1960, "Mozart"),
			Reference:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			WantPass:     false,
			WantErrors:   2, // Composer + title
			WantWarnings: 1, // Year
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.TagAccuracyVsReference(tt.Actual, tt.Reference)

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

func TestGetComposer(t *testing.T) {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	soloist := domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

	tests := []struct {
		Name    string
		Artists []domain.Artist
		Want    string
	}{
		{
			Name:    "has composer",
			Artists: []domain.Artist{composer, soloist, ensemble},
			Want:    "Beethoven",
		},
		{
			Name:    "no composer",
			Artists: []domain.Artist{soloist, ensemble},
			Want:    "",
		},
		{
			Name:    "empty list",
			Artists: []domain.Artist{},
			Want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := getComposer(tt.Artists)
			if got != tt.Want {
				t.Errorf("getComposer() = %q, want %q", got, tt.Want)
			}
		})
	}
}

// buildBasicAlbum creates a simple album for testing
func buildBasicAlbum(trackTitle string, year int, composerName string) *domain.Album {
	composer := domain.Artist{Name: composerName, Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	track := &domain.Track{Disc: 1, Track: 1, Title: trackTitle, Artists: []domain.Artist{composer, ensemble}}
	return &domain.Album{Title: "Album", OriginalYear: year, Tracks: []*domain.Track{track}}
}
