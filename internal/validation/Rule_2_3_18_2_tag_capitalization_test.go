package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TagCapitalization(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantErrors int
	}{
		{
			name:     "valid - Title Case",
			actual:   buildAlbumWithTitle("Symphony No. 5 in C Minor", "1963"),
			wantPass: true,
		},
		{
			name:     "valid - Casual Title Case",
			actual:   buildAlbumWithTitle("Symphony No. 5 In C Minor", "1963"),
			wantPass: true,
		},
		{
			name:       "error - all uppercase",
			actual:     buildAlbumWithTitle("SYMPHONY NO. 5", "1963"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - all lowercase",
			actual:     buildAlbumWithTitle("symphony no. 5", "1963"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:     "valid - mixed case acceptable",
			actual:   buildAlbumWithTitle("Symphony No. 5", "1963"),
			wantPass: true,
		},
		{
			name:       "error - artist name all caps",
			actual:     buildAlbumWithArtistName("BEETHOVEN"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - artist name all lowercase",
			actual:     buildAlbumWithArtistName("beethoven"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:     "valid - proper artist name",
			actual:   buildAlbumWithArtistName("Ludwig van Beethoven"),
			wantPass: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.TagCapitalization(tt.actual, tt.actual)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass {
				errorCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelError {
						errorCount++
					}
				}
				
				if errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}
				
				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

func TestRules_TagCapitalizationVsReference(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name         string
		actual       *domain.Album
		reference    *domain.Album
		wantPass     bool
		wantWarnings int
	}{
		{
			name:      "valid - exact match",
			actual:    buildAlbumWithTitle("Symphony No. 5", "1963"),
			reference: buildAlbumWithTitle("Symphony No. 5", "1963"),
			wantPass:  true,
		},
		{
			name:         "warning - case mismatch",
			actual:       buildAlbumWithTitle("Symphony No. 5", "1963"),
			reference:    buildAlbumWithTitle("Symphony no. 5", "1963"),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - all caps vs proper",
			actual:       buildAlbumWithTitle("SYMPHONY NO. 5", "1963"),
			reference:    buildAlbumWithTitle("Symphony No. 5", "1963"),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:      "pass - no reference",
			actual:    buildAlbumWithTitle("Symphony No. 5", "1963"),
			reference: nil,
			wantPass:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.TagCapitalizationVsReference(tt.actual, tt.reference)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass {
				warningCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelWarning {
						warningCount++
					}
				}
				
				if warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}
				
				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
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
