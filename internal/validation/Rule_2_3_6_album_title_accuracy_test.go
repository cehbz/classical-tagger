package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_AlbumTitleAccuracy(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		ActualTitle  string
		RefTitle     string
		WantPass     bool
		WantErrors   int
		WantWarnings int
	}{
		{
			Name:        "valid - exact match",
			ActualTitle: "Symphony No. 5 in C Minor",
			RefTitle:    "Symphony No. 5 in C Minor",
			WantPass:    true,
		},
		{
			Name:        "valid - minor punctuation differences",
			ActualTitle: "Symphony No. 5 in C Minor",
			RefTitle:    "Symphony No 5 in C Minor",
			WantPass:    true,
		},
		{
			Name:        "valid - case differences",
			ActualTitle: "Symphony No. 5",
			RefTitle:    "SYMPHONY NO. 5",
			WantPass:    true,
		},
		{
			Name:        "valid - minor typo (edit distance ≤3)",
			ActualTitle: "Sympony No. 5",
			RefTitle:    "Symphony No. 5",
			WantPass:    true,
		},
		{
			Name:         "warning - moderate differences (4-10 chars)",
			ActualTitle:  "Symphony No. 5 in C",
			RefTitle:     "Symphony No. 5 in C Minor, Op. 67",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:        "error - major differences (>10 chars)",
			ActualTitle: "Piano Concerto No. 1",
			RefTitle:    "Symphony No. 5 in C Minor",
			WantPass:    false,
			WantErrors:  1,
		},
		{
			Name:        "valid - abbreviation vs full title",
			ActualTitle: "Symphony No. 5",
			RefTitle:    "Symphony No. 5 in C Minor, Op. 67",
			WantPass:    true, // Substring match OK
		},
		{
			Name:        "valid - additional bracketed info",
			ActualTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			RefTitle:    "Beethoven - Symphony No. 5",
			WantPass:    true, // Core title matches
		},
		{
			Name:        "error - wrong work",
			ActualTitle: "Concerto in D Major",
			RefTitle:    "Concerto in E Major",
			WantPass:    false,
			WantErrors:  1,
		},
		{
			Name:        "valid - with composer prefix",
			ActualTitle: "Beethoven: Symphony No. 5",
			RefTitle:    "Symphony No. 5",
			WantPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			actual := NewAlbum().WithTitle(tt.ActualTitle).WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build()
			reference := NewAlbum().WithTitle(tt.RefTitle).WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build()

			result := rules.AlbumTitleAccuracy(actual, reference)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelError {
						errorCount++
					} else if issue.Level == domain.LevelWarning {
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

func TestRules_AlbumTitleAccuracy_NoReference(t *testing.T) {
	rules := NewRules()

	actual := NewAlbum().WithTitle("Any Title").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build()
	result := rules.AlbumTitleAccuracy(actual, nil)

	if !result.Passed() {
		t.Error("Should pass when no reference provided")
	}
}
