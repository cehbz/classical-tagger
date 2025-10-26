package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_RecordingDateVsYear(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name      string
		actual    *domain.Album
		reference *domain.Album
		wantPass  bool
		wantInfo  int
	}{
		{
			name:     "pass - no edition year",
			actual:   buildAlbumWithTitle("Symphony", "No. 9"),
			wantPass: true,
		},
		{
			name:     "pass - same year",
			actual:   buildAlbumWithEditionYear(1963, 1963),
			wantPass: true,
		},
		{
			name:     "pass - small difference (1-2 years)",
			actual:   buildAlbumWithEditionYear(1963, 1965),
			wantPass: true,
		},
		{
			name:     "info - moderate gap (3-10 years)",
			actual:   buildAlbumWithEditionYear(1963, 1968),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - large gap (>10 years)",
			actual:   buildAlbumWithEditionYear(1963, 1990),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - edition before recording",
			actual:   buildAlbumWithEditionYear(1990, 1985),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:      "info - differs from reference",
			actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			reference: buildAlbumWithTitle("Symphony", "Number nine"),
			wantPass:  false,
			wantInfo:  1,
		},
		{
			name:      "pass - close to reference",
			actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			reference: buildAlbumWithTitle("Symphony", "No 9"),
			wantPass:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.RecordingDateVsYear(tt.actual, tt.reference)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass {
				infoCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelInfo {
						infoCount++
					}
				}
				
				if infoCount != tt.wantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.wantInfo)
				}
				
				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}
