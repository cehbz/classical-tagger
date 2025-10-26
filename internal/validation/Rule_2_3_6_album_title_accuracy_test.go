package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_AlbumTitleAccuracy(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name         string
		actualTitle  string
		refTitle     string
		wantPass     bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name:        "valid - exact match",
			actualTitle: "Symphony No. 5 in C Minor",
			refTitle:    "Symphony No. 5 in C Minor",
			wantPass:    true,
		},
		{
			name:        "valid - minor punctuation differences",
			actualTitle: "Symphony No. 5 in C Minor",
			refTitle:    "Symphony No 5 in C Minor",
			wantPass:    true,
		},
		{
			name:        "valid - case differences",
			actualTitle: "Symphony No. 5",
			refTitle:    "SYMPHONY NO. 5",
			wantPass:    true,
		},
		{
			name:        "valid - minor typo (edit distance ≤3)",
			actualTitle: "Sympony No. 5",
			refTitle:    "Symphony No. 5",
			wantPass:    true,
		},
		{
			name:         "warning - moderate differences (4-10 chars)",
			actualTitle:  "Symphony No. 5 in C",
			refTitle:     "Symphony No. 5 in C Minor, Op. 67",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "error - major differences (>10 chars)",
			actualTitle: "Piano Concerto No. 1",
			refTitle:   "Symphony No. 5 in C Minor",
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:        "valid - abbreviation vs full title",
			actualTitle: "Symphony No. 5",
			refTitle:    "Symphony No. 5 in C Minor, Op. 67",
			wantPass:    true, // Substring match OK
		},
		{
			name:        "valid - additional bracketed info",
			actualTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			refTitle:    "Beethoven - Symphony No. 5",
			wantPass:    true, // Core title matches
		},
		{
			name:       "error - wrong work",
			actualTitle: "Concerto in D Major",
			refTitle:   "Concerto in E Major",
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:        "valid - with composer prefix",
			actualTitle: "Beethoven: Symphony No. 5",
			refTitle:    "Symphony No. 5",
			wantPass:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildAlbumWithTitle(tt.actualTitle, "1963")
			reference := buildAlbumWithTitle(tt.refTitle, "1963")
			
			result := rules.AlbumTitleAccuracy(actual, reference)
			
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

func TestRules_AlbumTitleAccuracy_NoReference(t *testing.T) {
	rules := NewRules()
	
	actual := buildAlbumWithTitle("Any Title", "1963")
	result := rules.AlbumTitleAccuracy(actual, nil)
	
	if !result.Passed() {
		t.Error("Should pass when no reference provided")
	}
}
