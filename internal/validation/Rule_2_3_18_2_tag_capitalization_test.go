package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TagCapitalization(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantErrors int
	}{
		{
			Name:     "valid - Title Case",
			Actual:   buildAlbumWithTitle("Symphony No. 5 in C Minor", "1963"),
			WantPass: true,
		},
		{
			Name:     "valid - Casual Title Case",
			Actual:   buildAlbumWithTitle("Symphony No. 5 In C Minor", "1963"),
			WantPass: true,
		},
		{
			Name:       "error - all uppercase",
			Actual:     buildAlbumWithTitle("SYMPHONY NO. 5", "1963"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - all lowercase",
			Actual:     buildAlbumWithTitle("symphony no. 5", "1963"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - mixed case acceptable",
			Actual:   buildAlbumWithTitle("Symphony No. 5", "1963"),
			WantPass: true,
		},
		{
			Name:       "error - artist name all caps",
			Actual:     buildAlbumWithArtistName("BEETHOVEN"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - artist name all lowercase",
			Actual:     buildAlbumWithArtistName("beethoven"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - proper artist name",
			Actual:   buildAlbumWithArtistName("Ludwig van Beethoven"),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.TagCapitalization(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelError {
						errorCount++
					}
				}

				if errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

func TestRules_TagCapitalizationVsReference(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		Reference    *domain.Album
		WantPass     bool
		WantWarnings int
	}{
		{
			Name:      "valid - exact match",
			Actual:    buildAlbumWithTitle("Symphony No. 5", "1963"),
			Reference: buildAlbumWithTitle("Symphony No. 5", "1963"),
			WantPass:  true,
		},
		{
			Name:         "warning - case mismatch",
			Actual:       buildAlbumWithTitle("Symphony No. 5", "1963"),
			Reference:    buildAlbumWithTitle("Symphony no. 5", "1963"),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - all caps vs proper",
			Actual:       buildAlbumWithTitle("SYMPHONY NO. 5", "1963"),
			Reference:    buildAlbumWithTitle("Symphony No. 5", "1963"),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:      "pass - no reference",
			Actual:    buildAlbumWithTitle("Symphony No. 5", "1963"),
			Reference: nil,
			WantPass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.TagCapitalizationVsReference(tt.Actual, tt.Reference)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				warningCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelWarning {
						warningCount++
					}
				}

				if warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

func TestCapitalizationMatches(t *testing.T) {
	tests := []struct {
		s1   string
		s2   string
		want bool
	}{
		{"Symphony No. 5", "Symphony No. 5", true},
		{"Symphony No. 5", "Symphony no. 5", false},
		{"Symphony No. 5", "SYMPHONY NO. 5", false},
		{"Test", "Test", true},
		{"test", "TEST", false},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_vs_"+tt.s2, func(t *testing.T) {
			got := capitalizationMatches(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("capitalizationMatches(%q, %q) = %v, want %v",
					tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}
