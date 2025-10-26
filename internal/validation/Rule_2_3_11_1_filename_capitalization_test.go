package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FilenameCapitalization(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - Title Case",
			actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5 in C Minor.flac",
				"02 - Concerto for Violin.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - Casual Title Case (every word capitalized)",
			actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5 In C Minor.flac",
				"02 - Concerto For Violin And Orchestra.flac",
			),
			wantPass: true,
		},
		{
			name: "invalid - all uppercase",
			actual: buildAlbumWithFilenames(
				"01 - SYMPHONY NO. 5.flac",
				"02 - CONCERTO.flac",
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "invalid - all lowercase",
			actual: buildAlbumWithFilenames(
				"01 - symphony no. 5.flac",
				"02 - concerto.flac",
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "invalid - some tracks all caps",
			actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5.flac",
				"02 - CONCERTO IN D.flac",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "valid - mixed case (acceptable)",
			actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5.flac",
				"02 - Concerto in D major.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - with numbers and abbreviations",
			actual: buildAlbumWithFilenames(
				"01 - BWV 1007 - Prelude.flac",
				"02 - Op. 132 - Allegro.flac",
			),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.FilenameCapitalization(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass && len(result.Issues()) != tt.wantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues()), tt.wantIssues)
				for _, issue := range result.Issues() {
					t.Logf("  Issue: %s", issue.Message())
				}
			}
		})
	}
}

func TestCheckCapitalization(t *testing.T) {
	tests := []struct {
		title   string
		wantErr bool
		errMsg  string
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
		t.Run(tt.title, func(t *testing.T) {
			got := checkCapitalization(tt.title)
			hasErr := got != ""

			if hasErr != tt.wantErr {
				t.Errorf("checkCapitalization(%q) error = %v, want %v", tt.title, hasErr, tt.wantErr)
			}

			if tt.wantErr && got != tt.errMsg {
				t.Errorf("checkCapitalization(%q) = %q, want %q", tt.title, got, tt.errMsg)
			}
		})
	}
}

func TestAdditionalCapitalizationCases(t *testing.T) {
	cases := []struct {
		title string
		ok    bool
	}{
		{"Well-Tempered Clavier", true},
		{"Symphony No. 5: Allegro con brio", true},
		{"LSO – Symphony No. II", true},
		{"R&B Anthology", true},
		{"Messa da Requiem", true},
		{"Concerto per pianoforte", true},
	}
	for _, c := range cases {
		got := checkCapitalization(c.title)
		if (got == "") != c.ok {
			t.Errorf("%q ok=%v, got=%q", c.title, c.ok, got)
		}
	}
}

func TestIsAllUppercase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
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
		t.Run(tt.input, func(t *testing.T) {
			got := isAllUppercase(tt.input)
			if got != tt.want {
				t.Errorf("isAllUppercase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsAllLowercase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
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
		t.Run(tt.input, func(t *testing.T) {
			got := isAllLowercase(tt.input)
			if got != tt.want {
				t.Errorf("isAllLowercase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsTitleCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
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
		t.Run(tt.input, func(t *testing.T) {
			got := isTitleCase(tt.input)
			if got != tt.want {
				t.Errorf("isTitleCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
