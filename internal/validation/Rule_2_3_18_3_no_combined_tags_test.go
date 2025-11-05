package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoCombinedTags(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Torrent
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:     "valid - separate artist entries",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Maurizio Pollini", Role: domain.RoleSoloist}, domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:         "warning - combined artist names with semicolon",
			Actual:       NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("Pollini; Arrau", domain.RoleSoloist).Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - combined artist names with slash",
			Actual:       NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("Martha Argerich / Nelson Freire", domain.RoleSoloist).Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - combined artist names with ampersand",
			Actual:       NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("Perlman & Ashkenazy", domain.RoleSoloist).Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:     "valid - ensemble name with 'and'",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("London Symphony Orchestra and Chorus", domain.RoleEnsemble).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - orchestra name with 'of'",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("Orchestra of the Age of Enlightenment", domain.RoleEnsemble).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - quartet name",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("Emerson String Quartet", domain.RoleEnsemble).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - compound last name",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("Mendelssohn-Bartholdy", domain.RoleComposer).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "info - combined works in title",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Symphony No. 1 / Symphony No. 2").Build().Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - movement subtitle with slash",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Allegro / Fast").Build().Build(),
			WantPass: true, // Short parts, not multiple works
		},
	}

	for _, tt := range tests {
		for _, track := range tt.Actual.Tracks() {
			t.Run(tt.Name, func(t *testing.T) {
				result := rules.NoCombinedTags(track, nil, nil, nil)

				if result.Passed() != tt.WantPass {
					t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
				}

				if !tt.WantPass {
					warningCount := 0
					infoCount := 0
					for _, issue := range result.Issues {
						switch issue.Level {
						case domain.LevelWarning:
							warningCount++
						case domain.LevelInfo:
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
