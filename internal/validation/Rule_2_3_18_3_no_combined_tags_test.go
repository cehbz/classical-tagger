package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoCombinedTags(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name: "valid - separate artist entries",
			Actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Maurizio Pollini", domain.RoleSoloist,
				"Berlin Philharmonic", domain.RoleEnsemble,
			),
			WantPass: true,
		},
		{
			Name:         "warning - combined artist names with semicolon",
			Actual:       buildAlbumWithSingleArtist("Pollini; Arrau", domain.RoleSoloist),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - combined artist names with slash",
			Actual:       buildAlbumWithSingleArtist("Martha Argerich / Nelson Freire", domain.RoleSoloist),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - combined artist names with ampersand",
			Actual:       buildAlbumWithSingleArtist("Perlman & Ashkenazy", domain.RoleSoloist),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:     "valid - ensemble name with 'and'",
			Actual:   buildAlbumWithSingleArtist("London Symphony Orchestra and Chorus", domain.RoleEnsemble),
			WantPass: true,
		},
		{
			Name:     "valid - orchestra name with 'of'",
			Actual:   buildAlbumWithSingleArtist("Orchestra of the Age of Enlightenment", domain.RoleEnsemble),
			WantPass: true,
		},
		{
			Name:     "valid - quartet name",
			Actual:   buildAlbumWithSingleArtist("Emerson String Quartet", domain.RoleEnsemble),
			WantPass: true,
		},
		{
			Name:     "valid - compound last name",
			Actual:   buildAlbumWithSingleArtist("Mendelssohn-Bartholdy", domain.RoleComposer),
			WantPass: true,
		},
		{
			Name:     "info - combined works in title",
			Actual:   buildAlbumWithTrackTitle("Symphony No. 1 / Symphony No. 2"),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - movement subtitle with slash",
			Actual:   buildAlbumWithTrackTitle("Allegro / Fast"),
			WantPass: true, // Short parts, not multiple works
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.NoCombinedTags(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelWarning {
						warningCount++
					} else if issue.Level == domain.LevelInfo {
						infoCount++
					}
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

func TestIsMultipleArtists(t *testing.T) {
	tests := []struct {
		Name       string
		ArtistName string
		Separator  string
		Want       bool
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
		t.Run(tt.Name, func(t *testing.T) {
			got := isMultipleArtists(tt.ArtistName, tt.Separator)
			if got != tt.Want {
				t.Errorf("isMultipleArtists(%q, %q) = %v, want %v",
					tt.ArtistName, tt.Separator, got, tt.Want)
			}
		})
	}
}

// buildAlbumWithSingleArtist creates album with one artist having a specific name
func buildAlbumWithSingleArtist(artistName string, role domain.Role) *domain.Album {
	artist := domain.Artist{Name: artistName, Role: role}
	track := domain.Track{Disc: 1, Track: 1, Title: "Work", Artists: []domain.Artist{artist}}
	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: []*domain.Track{&track}}
}

// buildAlbumWithTrackTitle creates album with specific track title
func buildAlbumWithTrackTitle(title string) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	track := &domain.Track{Disc: 1, Track: 1, Title: title, Artists: []domain.Artist{composer, ensemble}}
	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: []*domain.Track{track}}
}
