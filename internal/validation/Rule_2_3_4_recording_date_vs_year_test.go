package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_RecordingDateVsYear(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name      string
		Actual    *domain.Album
		Reference *domain.Album
		WantPass  bool
		WantInfo  int
	}{
		{
			Name:     "pass - no edition year",
			Actual:   buildAlbumWithTitle("Symphony", "No. 9"),
			WantPass: true,
		},
		{
			Name:     "pass - same year",
			Actual:   buildAlbumWithEditionYear(1963, 1963),
			WantPass: true,
		},
		{
			Name:     "pass - small difference (1-2 years)",
			Actual:   buildAlbumWithEditionYear(1963, 1965),
			WantPass: true,
		},
		{
			Name:     "info - moderate gap (3-10 years)",
			Actual:   buildAlbumWithEditionYear(1963, 1968),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - large gap (>10 years)",
			Actual:   buildAlbumWithEditionYear(1963, 1990),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - edition before recording",
			Actual:   buildAlbumWithEditionYear(1990, 1985),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:      "info - differs from reference",
			Actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			Reference: buildAlbumWithTitle("Symphony", "Number nine"),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "pass - close to reference",
			Actual:    buildAlbumWithTitle("Symphony", "No. 9"),
			Reference: buildAlbumWithTitle("Symphony", "No 9"),
			WantPass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.RecordingDateVsYear(tt.Actual, tt.Reference)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				infoCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelInfo {
						infoCount++
					}
				}

				if infoCount != tt.WantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}
