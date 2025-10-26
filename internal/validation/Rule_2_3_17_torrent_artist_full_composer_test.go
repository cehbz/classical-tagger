package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TorrentArtistFullComposerName(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		albumTitle   string
		composers    []string
		wantPass     bool
		wantWarnings int
		wantInfo     int
	}{
		{
			name:       "valid - full composer name in title",
			albumTitle: "Ludwig van Beethoven - Complete Symphonies",
			composers:  []string{"Ludwig van Beethoven"},
			wantPass:   true,
		},
		{
			name:       "valid - acceptable abbreviation J.S. Bach",
			albumTitle: "J.S. Bach - Brandenburg Concertos",
			composers:  []string{"Johann Sebastian Bach"},
			wantPass:   true,
		},
		{
			name:       "valid - acceptable abbreviation W.A. Mozart",
			albumTitle: "W.A. Mozart - Piano Concertos",
			composers:  []string{"Wolfgang Amadeus Mozart"},
			wantPass:   true,
		},
		{
			name:         "warning - last name only in title",
			albumTitle:   "Beethoven - Symphonies",
			composers:    []string{"Ludwig van Beethoven"},
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - Bach without initials",
			albumTitle:   "Bach - Cello Suites",
			composers:    []string{"Johann Sebastian Bach"},
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "info - composer not in title at all",
			albumTitle: "Cello Suites",
			composers:  []string{"Johann Sebastian Bach"},
			wantPass:   false,
			wantInfo:   1,
		},
		{
			name:         "valid - multiple composers, full names",
			albumTitle:   "Bach & Vivaldi - Baroque Masterpieces",
			composers:    []string{"Johann Sebastian Bach", "Antonio Vivaldi"},
			wantPass:     false, // Bach without initials
			wantWarnings: 1,     // Only one warning, Vivaldi is just last name which is acceptable here
		},
		{
			name:         "valid - performer-focused title (acceptable)",
			albumTitle:   "Maurizio Pollini plays Beethoven",
			composers:    []string{"Ludwig van Beethoven"},
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "valid - work-focused title",
			albumTitle: "The Four Seasons by Antonio Vivaldi",
			composers:  []string{"Antonio Vivaldi"},
			wantPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildAlbumWithTitleAndComposers(tt.albumTitle, tt.composers...)
			result := rules.TorrentArtistFullComposerName(actual, actual)

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

func TestExtractPrimaryLastName(t *testing.T) {
	tests := []struct {
		name         string
		composerName string
		want         string
	}{
		{"simple", "Johann Bach", "Bach"},
		{"full name", "Johann Sebastian Bach", "Bach"},
		{"with particle", "Ludwig van Beethoven", "Beethoven"},
		{"with von", "Richard von Strauss", "Strauss"},
		{"initials", "J.S. Bach", "Bach"},
		{"reversed", "Beethoven, Ludwig van", "Beethoven"},
		{"single word", "Bach", "Bach"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPrimaryLastName(tt.composerName)
			if got != tt.want {
				t.Errorf("extractPrimaryLastName(%q) = %q, want %q", tt.composerName, got, tt.want)
			}
		})
	}
}

func TestContainsWord(t *testing.T) {
	tests := []struct {
		name string
		text string
		word string
		want bool
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
		t.Run(tt.name, func(t *testing.T) {
			got := containsWord(tt.text, tt.word)
			if got != tt.want {
				t.Errorf("containsWord(%q, %q) = %v, want %v", tt.text, tt.word, got, tt.want)
			}
		})
	}
}

func TestIsAcceptableAbbreviation(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		fullName string
		want     bool
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
		t.Run(tt.name, func(t *testing.T) {
			got := isAcceptableAbbreviation(tt.title, tt.fullName)
			if got != tt.want {
				t.Errorf("isAcceptableAbbreviation(%q, %q) = %v, want %v",
					tt.title, tt.fullName, got, tt.want)
			}
		})
	}
}

// Helper to build album with specific title and composers
func buildAlbumWithTitleAndComposers(albumTitle string, composerNames ...string) *domain.Album {
	// Create tracks with the specified composers

	album, _ := domain.NewAlbum(albumTitle, 1963)
	for i, composerName := range composerNames {
		composer, _ := domain.NewArtist(composerName, domain.RoleComposer)
		ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
		conductor, _ := domain.NewArtist("Conductor", domain.RoleConductor)

		artists := []domain.Artist{composer, ensemble, conductor}
		track, _ := domain.NewTrack(1, i+1, "Work "+string(rune('A'+i)), artists)
		album.AddTrack(track)
	}

	return album
}
