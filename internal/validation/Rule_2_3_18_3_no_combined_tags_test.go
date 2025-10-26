package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoCombinedTags(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		actual       *domain.Album
		wantPass     bool
		wantWarnings int
		wantInfo     int
	}{
		{
			name: "valid - separate artist entries",
			actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Maurizio Pollini", domain.RoleSoloist,
				"Berlin Philharmonic", domain.RoleEnsemble,
			),
			wantPass: true,
		},
		{
			name:         "warning - combined artist names with semicolon",
			actual:       buildAlbumWithSingleArtist("Pollini; Arrau", domain.RoleSoloist),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - combined artist names with slash",
			actual:       buildAlbumWithSingleArtist("Martha Argerich / Nelson Freire", domain.RoleSoloist),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - combined artist names with ampersand",
			actual:       buildAlbumWithSingleArtist("Perlman & Ashkenazy", domain.RoleSoloist),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:     "valid - ensemble name with 'and'",
			actual:   buildAlbumWithSingleArtist("London Symphony Orchestra and Chorus", domain.RoleEnsemble),
			wantPass: true,
		},
		{
			name:     "valid - orchestra name with 'of'",
			actual:   buildAlbumWithSingleArtist("Orchestra of the Age of Enlightenment", domain.RoleEnsemble),
			wantPass: true,
		},
		{
			name:     "valid - quartet name",
			actual:   buildAlbumWithSingleArtist("Emerson String Quartet", domain.RoleEnsemble),
			wantPass: true,
		},
		{
			name:     "valid - compound last name",
			actual:   buildAlbumWithSingleArtist("Mendelssohn-Bartholdy", domain.RoleComposer),
			wantPass: true,
		},
		{
			name:     "info - combined works in title",
			actual:   buildAlbumWithTrackTitle("Symphony No. 1 / Symphony No. 2"),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "valid - movement subtitle with slash",
			actual:   buildAlbumWithTrackTitle("Allegro / Fast"),
			wantPass: true, // Short parts, not multiple works
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.NoCombinedTags(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelWarning {
						warningCount++
					} else if issue.Level() == domain.LevelInfo {
						infoCount++
					}
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

func TestIsMultipleArtists(t *testing.T) {
	tests := []struct {
		name       string
		artistName string
		separator  string
		want       bool
	}{
		{"multiple with semicolon", "Pollini; Arrau", ";", true},
		{"multiple with slash", "Martha Argerich / Nelson Freire", " / ", true},
		{"orchestra with 'and'", "London Symphony Orchestra and Chorus", " and ", false},
		{"orchestra with 'of'", "Orchestra of the Age of Enlightenment", " of ", false},
		{"compound last name", "Mendelssohn-Bartholdy", ", ", false},
		{"quartet name", "Emerson String Quartet", " & ", false},
		{"two soloists", "Anne-Sophie Mutter & Yo-Yo Ma", " & ", true},
		{"initials only", "J.S. & C.P.E.", " & ", false}, // Too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMultipleArtists(tt.artistName, tt.separator)
			if got != tt.want {
				t.Errorf("isMultipleArtists(%q, %q) = %v, want %v",
					tt.artistName, tt.separator, got, tt.want)
			}
		})
	}
}

// buildAlbumWithSingleArtist creates album with one artist having a specific name
func buildAlbumWithSingleArtist(artistName string, role domain.Role) *domain.Album {
	artist, _ := domain.NewArtist(artistName, role)
	track, _ := domain.NewTrack(1, 1, "Work", []domain.Artist{artist})
	album, _ := domain.NewAlbum("Album", 1963)
	album.AddTrack(track)
	return album
}

// buildAlbumWithTrackTitle creates album with specific track title
func buildAlbumWithTrackTitle(title string) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, title, []domain.Artist{composer, ensemble})
	album, _ := domain.NewAlbum("Album", 1963)
	album.AddTrack(track)
	return album
}
