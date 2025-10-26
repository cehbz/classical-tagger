package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TagAccuracyVsReference(t *testing.T) {
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
			name:      "valid - exact match",
			actual:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			reference: buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			wantPass:  true,
		},
		{
			name:         "warning - year mismatch",
			actual:       buildBasicAlbum("Symphony No. 5", 1960, "Beethoven"),
			reference:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "error - composer mismatch",
			actual:     buildBasicAlbum("Symphony No. 5", 1963, "Mozart"),
			reference:  buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - title mismatch",
			actual:     buildBasicAlbum("Symphony No. 6", 1963, "Beethoven"),
			reference:  buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:      "info - minor title difference",
			actual:    buildBasicAlbum("Sympony No. 5", 1963, "Beethoven"), // Typo
			reference: buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			wantPass:  false,
			wantInfo:  1,
		},
		{
			name:         "warning - moderate title difference",
			actual:       buildBasicAlbum("Symphony No. 5 Finale", 1963, "Beethoven"),
			reference:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:      "pass - no reference",
			actual:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			reference: nil,
			wantPass:  true,
		},
		{
			name:         "multiple errors",
			actual:       buildBasicAlbum("Concerto", 1960, "Mozart"),
			reference:    buildBasicAlbum("Symphony No. 5", 1963, "Beethoven"),
			wantPass:     false,
			wantErrors:   2, // Composer + title
			wantWarnings: 1, // Year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.TagAccuracyVsReference(tt.actual, tt.reference)

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

func TestGetComposer(t *testing.T) {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	soloist, _ := domain.NewArtist("Pollini", domain.RoleSoloist)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)

	tests := []struct {
		name    string
		artists []domain.Artist
		want    string
	}{
		{
			name:    "has composer",
			artists: []domain.Artist{composer, soloist, ensemble},
			want:    "Beethoven",
		},
		{
			name:    "no composer",
			artists: []domain.Artist{soloist, ensemble},
			want:    "",
		},
		{
			name:    "empty list",
			artists: []domain.Artist{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getComposer(tt.artists)
			if got != tt.want {
				t.Errorf("getComposer() = %q, want %q", got, tt.want)
			}
		})
	}
}

// buildBasicAlbum creates a simple album for testing
func buildBasicAlbum(trackTitle string, year int, composerName string) *domain.Album {
	composer, _ := domain.NewArtist(composerName, domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, trackTitle, []domain.Artist{composer, ensemble})
	album, _ := domain.NewAlbum("Album", year)
	album.AddTrack(track)
	return album
}
