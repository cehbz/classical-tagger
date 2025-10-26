package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoRequestTagInTitle(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name         string
		albumTitle   string
		wantPass     bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name:       "valid - no request tag",
			albumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			wantPass:   true,
		},
		{
			name:       "invalid - [REQ] tag present",
			albumTitle: "Beethoven - Symphony No. 5 [REQ] [1963] [FLAC]",
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "invalid - [REQ] lowercase",
			albumTitle: "Beethoven - Symphony No. 5 [req] [1963] [FLAC]",
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "invalid - [REQ] mixed case",
			albumTitle: "Beethoven - Symphony No. 5 [Req] [1963] [FLAC]",
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:         "warning - [REQUEST] tag",
			albumTitle:   "Beethoven - Symphony No. 5 [REQUEST] [1963] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - [REQUESTED] tag",
			albumTitle:   "Beethoven - Symphony No. 5 [REQUESTED] [1963] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - (REQ) variant",
			albumTitle:   "Beethoven - Symphony No. 5 (REQ) [1963] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - (REQUEST) variant",
			albumTitle:   "Beethoven - Symphony No. 5 (REQUEST) [1963] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "valid - 'Required' in normal context",
			albumTitle: "Bach - Required Listening [1990] [FLAC]",
			wantPass:   true, // "Required" as a word is OK, not [REQ] tag
		},
		{
			name:       "valid - album with numbers",
			albumTitle: "Mozart - Requiem K.626 [1991] [FLAC]",
			wantPass:   true, // "Requiem" contains "req" but not [REQ] tag
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildAlbumWithTitle(tt.albumTitle, "1963")
			result := rules.NoRequestTagInTitle(actual, actual)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelError {
						errorCount++
					} else if issue.Level() == domain.LevelWarning {
						warningCount++
					}
				}
				
				if tt.wantErrors > 0 && errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}
				if tt.wantWarnings > 0 && warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}
				
				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}
