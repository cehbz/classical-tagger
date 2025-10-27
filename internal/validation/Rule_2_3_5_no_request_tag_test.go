package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoRequestTagInTitle(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		AlbumTitle   string
		WantPass     bool
		WantErrors   int
		WantWarnings int
	}{
		{
			Name:       "valid - no request tag",
			AlbumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			WantPass:   true,
		},
		{
			Name:       "invalid - [REQ] tag present",
			AlbumTitle: "Beethoven - Symphony No. 5 [REQ] [1963] [FLAC]",
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "invalid - [REQ] lowercase",
			AlbumTitle: "Beethoven - Symphony No. 5 [req] [1963] [FLAC]",
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "invalid - [REQ] mixed case",
			AlbumTitle: "Beethoven - Symphony No. 5 [Req] [1963] [FLAC]",
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:         "warning - [REQUEST] tag",
			AlbumTitle:   "Beethoven - Symphony No. 5 [REQUEST] [1963] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - [REQUESTED] tag",
			AlbumTitle:   "Beethoven - Symphony No. 5 [REQUESTED] [1963] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - (REQ) variant",
			AlbumTitle:   "Beethoven - Symphony No. 5 (REQ) [1963] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - (REQUEST) variant",
			AlbumTitle:   "Beethoven - Symphony No. 5 (REQUEST) [1963] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "valid - 'Required' in normal context",
			AlbumTitle: "Bach - Required Listening [1990] [FLAC]",
			WantPass:   true, // "Required" as a word is OK, not [REQ] tag
		},
		{
			Name:       "valid - album with numbers",
			AlbumTitle: "Mozart - Requiem K.626 [1991] [FLAC]",
			WantPass:   true, // "Requiem" contains "req" but not [REQ] tag
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			actual := buildAlbumWithTitle(tt.AlbumTitle, "1963")
			result := rules.NoRequestTagInTitle(actual, actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues {
					switch issue.Level {
case domain.LevelError:
						errorCount++
					case domain.LevelWarning:
						warningCount++
					}
				}

				if tt.WantErrors > 0 && errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}
				if tt.WantWarnings > 0 && warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}
