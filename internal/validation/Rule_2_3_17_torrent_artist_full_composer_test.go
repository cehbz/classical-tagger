package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TorrentArtistFullComposerName(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		AlbumTitle   string
		Composers    []string
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:       "valid - full composer name in title",
			AlbumTitle: "Ludwig van Beethoven - Complete Symphonies",
			Composers:  []string{"Ludwig van Beethoven"},
			WantPass:   true,
		},
		{
			Name:       "valid - acceptable abbreviation J.S. Bach",
			AlbumTitle: "J.S. Bach - Brandenburg Concertos",
			Composers:  []string{"Johann Sebastian Bach"},
			WantPass:   true,
		},
		{
			Name:       "valid - acceptable abbreviation W.A. Mozart",
			AlbumTitle: "W.A. Mozart - Piano Concertos",
			Composers:  []string{"Wolfgang Amadeus Mozart"},
			WantPass:   true,
		},
		{
			Name:         "warning - last name only in title",
			AlbumTitle:   "Beethoven - Symphonies",
			Composers:    []string{"Ludwig van Beethoven"},
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - Bach without initials",
			AlbumTitle:   "Bach - Cello Suites",
			Composers:    []string{"Johann Sebastian Bach"},
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "info - composer not in title at all",
			AlbumTitle: "Cello Suites",
			Composers:  []string{"Johann Sebastian Bach"},
			WantPass:   false,
			WantInfo:   1,
		},
		{
			Name:       "valid - multiple composers, last names",
			AlbumTitle: "Bach & Vivaldi - Baroque Masterpieces",
			Composers:  []string{"Johann Sebastian Bach", "Antonio Vivaldi"},
			WantPass:   true, // No primary composer (>50% threshold), so rule doesn't apply
		},
		{
			Name:       "valid - one dominant composer, last name",
			AlbumTitle: "Vivaldi - Baroque Masterpieces",
			Composers:  []string{"Johann Sebastian Bach", "Antonio Vivaldi", "Antonio Vivaldi"},
			WantPass:   true,
		},
		{
			Name:       "info - one dominant composer, name missing",
			AlbumTitle: "Bach - Baroque Masterpieces",
			Composers:  []string{"Johann Sebastian Bach", "Antonio Vivaldi", "Antonio Vivaldi"},
			WantPass:   false,
			WantInfo:   1,
		},
		{
			Name:         "valid - performer-focused title (acceptable)",
			AlbumTitle:   "Maurizio Pollini plays Beethoven",
			Composers:    []string{"Ludwig van Beethoven"},
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "valid - work-focused title",
			AlbumTitle: "The Four Seasons by Antonio Vivaldi",
			Composers:  []string{"Antonio Vivaldi"},
			WantPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			builder := NewTorrent().WithTitle(tt.AlbumTitle).ClearTracks()
			for i, composerName := range tt.Composers {
				builder.AddTrack().
					WithTrack(i+1).
					WithTitle("Work "+string(rune('A'+i))).
					ClearArtists().
					WithArtists(
						domain.Artist{Name: composerName, Role: domain.RoleComposer},
						domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
						domain.Artist{Name: "Conductor", Role: domain.RoleConductor}).
					Build()
			}
			actual := builder.Build()
			result := rules.TorrentArtistFullComposerName(actual, nil)

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

func TestContainsWord(t *testing.T) {
	tests := []struct {
		Name string
		Text string
		Word string
		Want bool
	}{
		{"exact match", "Bach Cello Suites", "Bach", true},
		{"case insensitive", "bach cello suites", "Bach", true},
		{"not whole word", "Bacharach Suite", "Bach", false},
		{"with punctuation", "Bach: Cello Suites", "Bach", true},
		{"with dash", "J.S. Bach - Works", "Bach", true},
		{"not present", "Vivaldi Concertos", "Bach", false},
		{"abbreviation", "J.S. Bach", "J.S.", true},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := containsWord(tt.Text, tt.Word)
			if got != tt.Want {
				t.Errorf("containsWord(%q, %q) = %v, want %v", tt.Text, tt.Word, got, tt.Want)
			}
		})
	}
}

func TestIsAcceptableAbbreviation(t *testing.T) {
	tests := []struct {
		Name     string
		Title    string
		FullName string
		Want     bool
	}{
		{
			"J.S. Bach for Johann Sebastian Bach",
			"J.S. Bach - Brandenburg Concertos",
			"Johann Sebastian Bach",
			true,
		},
		{
			"W.A. Mozart for Wolfgang Amadeus Mozart",
			"W.A. Mozart - Piano Concertos",
			"Wolfgang Amadeus Mozart",
			true,
		},
		{
			"C.P.E. Bach for Carl Philipp Emanuel Bach",
			"C.P.E. Bach - Keyboard Concertos",
			"Carl Philipp Emanuel Bach",
			true,
		},
		{
			"no abbreviation present",
			"Bach - Cello Suites",
			"Johann Sebastian Bach",
			false,
		},
		{
			"single name - no abbreviation possible",
			"Bach Works",
			"Bach",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := isAcceptableAbbreviation(tt.Title, tt.FullName)
			if got != tt.Want {
				t.Errorf("isAcceptableAbbreviation(%q, %q) = %v, want %v",
					tt.Title, tt.FullName, got, tt.Want)
			}
		})
	}
}
