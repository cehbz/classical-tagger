package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FilenameCapitalization(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name:       "valid - Title Case",
			Actual:     buildAlbumWithFilenames("01 - Symphony No. 5 in C Minor.flac"),
			WantPass:   true,
			WantIssues: 0,
		},
		{
			Name:       "valid - Casual Title Case (every word capitalized)",
			Actual:     buildAlbumWithFilenames("01 - Symphony No. 5 In C Minor.flac"),
			WantPass:   true,
			WantIssues: 0,
		},
		{
			Name:       "invalid - all uppercase",
			Actual:     buildAlbumWithFilenames("01 - SYMPHONY NO. 5.flac"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - all lowercase",
			Actual:     buildAlbumWithFilenames("01 - symphony no. 5.flac"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "valid - mixed case (acceptable)",
			Actual:     buildAlbumWithFilenames("01 - Symphony No. 5.flac"),
			WantPass:   true,
			WantIssues: 0,
		},
		{
			Name:       "valid - with numbers and abbreviations",
			Actual:     buildAlbumWithFilenames("01 - BWV 1007 - Prelude.flac"),
			WantPass:   true,
			WantIssues: 0,
		},
	}

	for _, tt := range tests {
		for _, track := range tt.Actual.Tracks {
			t.Run(tt.Name, func(t *testing.T) {
				result := rules.FilenameCapitalization(track, nil, nil, nil)

				if result.Passed() != tt.WantPass {
					t.Errorf("FilenameCapitalization(%q) passed = %v, want %v", tt.Actual.Title, result.Passed(), tt.WantPass)
				}

				if !tt.WantPass && len(result.Issues) != tt.WantIssues {
					t.Errorf("Issues = %d, want %d", len(result.Issues), tt.WantIssues)
					for _, issue := range result.Issues {
						t.Logf("  Issue: %s", issue.Message)
					}
				}
			})
		}
	}
}

func TestCheckCapitalization(t *testing.T) {
	tests := []struct {
		Title   string
		WantErr bool
		ErrMsg  string
	}{
		{"Symphony No. 5", false, ""},
		{"SYMPHONY NO. 5", true, "Not Title Case or Casual Title Case"},
		{"symphony no. 5", true, "Not Title Case or Casual Title Case"},
		{"Symphony No. 5 in C Minor", false, ""}, // strict ok
		{"Symphony No. 5 In C Minor", false, ""}, // strict ok (allow In before key)
		{"BWV 1007", false, ""},                  // Abbreviations OK
		{"Op. 132", false, ""},
		{"Prelude and Fugue", false, ""},
		{"Symphony no. 5", true, "Not Title Case or Casual Title Case"}, // non-canonical 'no.'
		{"op. 132", true, "Not Title Case or Casual Title Case"},        // non-canonical 'op.'
		{"hob. XVI:52", true, "Not Title Case or Casual Title Case"},    // non-canonical 'hob.'
		{"bwv 988", true, "Not Title Case or Casual Title Case"},        // non-canonical BWV
	}

	for _, tt := range tests {
		t.Run(tt.Title, func(t *testing.T) {
			got := checkCapitalization(tt.Title)
			hasErr := got != ""

			if hasErr != tt.WantErr {
				t.Errorf("checkCapitalization(%q) error = %v, want %v", tt.Title, hasErr, tt.WantErr)
			}

			if tt.WantErr && got != tt.ErrMsg {
				t.Errorf("checkCapitalization(%q) = %q, want %q", tt.Title, got, tt.ErrMsg)
			}
		})
	}
}

func TestAdditionalCapitalizationCases(t *testing.T) {
	cases := []struct {
		Title string
		Ok    bool
	}{
		{"Well-Tempered Clavier", true},
		{"Symphony No. 5: Allegro con brio", true},
		{"LSO – Symphony No. II", true},
		{"R&B Anthology", true},
		{"Messa da Requiem", true},
		{"Concerto per pianoforte", true},
		{"RIAS Kammerchor", true}, // RIAS should be detected as acronym
		{"HMC 902170", true},      // HMC should be detected as acronym
		{"SYMPHONY NO. 5", false}, // SYMPHONY is too long to be an acronym
	}
	for _, c := range cases {
		got := checkCapitalization(c.Title)
		if (got == "") != c.Ok {
			t.Errorf("%q ok=%v, got=%q", c.Title, c.Ok, got)
		}
	}
}

func TestIsAllUppercase(t *testing.T) {
	tests := []struct {
		Input string
		Want  bool
	}{
		{"SYMPHONY", true},
		{"Symphony", false},
		{"SYMPHONY NO. 5", true},
		{"Symphony No. 5", false},
		{"123", false}, // No letters
		{"", false},
		{"ABC123", true}, // Letters are all uppercase
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			got := isAllUppercase(tt.Input)
			if got != tt.Want {
				t.Errorf("isAllUppercase(%q) = %v, want %v", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestIsAllLowercase(t *testing.T) {
	tests := []struct {
		Input string
		Want  bool
	}{
		{"symphony", true},
		{"Symphony", false},
		{"symphony no. 5", true},
		{"Symphony No. 5", false},
		{"123", false}, // No letters
		{"", false},
		{"abc123", true}, // Letters are all lowercase
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			got := isAllLowercase(tt.Input)
			if got != tt.Want {
				t.Errorf("isAllLowercase(%q) = %v, want %v", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestIsTitleCase(t *testing.T) {
	tests := []struct {
		Input string
		Want  bool
	}{
		{"Symphony No. 5", true},
		{"Symphony No. 5 in C Minor", true}, // "in" is ok lowercase
		{"symphony No. 5", false},
		{"Symphony no. 5", false},
		{"The Art of Fugue", true},  // "of" is ok lowercase
		{"A Prelude", true},         // "A" at start is capitalized
		{"Prelude and Fugue", true}, // "and" is ok lowercase
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			got := isTitleCase(tt.Input)
			if got != tt.Want {
				t.Errorf("isTitleCase(%q) = %v, want %v", tt.Input, got, tt.Want)
			}
		})
	}
}
