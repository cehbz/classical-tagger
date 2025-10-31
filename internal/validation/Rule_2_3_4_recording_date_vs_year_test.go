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
			Actual:   NewAlbum().WithTitle("Symphony").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "pass - same year",
			Actual:   NewAlbum().WithOriginalYear(1963).WithEdition("Label", "CAT123", 1963).Build(),
			WantPass: true,
		},
		{
			Name:     "pass - small difference (1-2 years)",
			Actual:   NewAlbum().WithOriginalYear(1963).WithEdition("Label", "CAT123", 1965).Build(),
			WantPass: true,
		},
		{
			Name:     "info - moderate gap (3-10 years)",
			Actual:   NewAlbum().WithOriginalYear(1963).WithEdition("Label", "CAT123", 1968).Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - large gap (>10 years)",
			Actual:   NewAlbum().WithOriginalYear(1963).WithEdition("Label", "CAT123", 1990).Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - edition before recording",
			Actual:   NewAlbum().WithOriginalYear(1990).WithEdition("Label", "CAT123", 1985).Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:      "info - differs from reference",
			Actual:    NewAlbum().WithOriginalYear(1963).WithEdition("Label", "CAT123", 1963).Build(),
			Reference: NewAlbum().WithOriginalYear(1990).WithEdition("Label", "CAT123", 1990).Build(),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "pass - close to reference",
			Actual:    NewAlbum().WithTitle("Symphony").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			Reference: NewAlbum().WithTitle("Symphony").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
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
