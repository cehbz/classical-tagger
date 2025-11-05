package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_YearFieldUsage(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Torrent
		Reference    *domain.Torrent
		WantPass     bool
		WantErrors   int
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:     "valid - reasonable year",
			Actual:   NewTorrent().WithTitle("Symphony").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:         "warning - year too early",
			Actual:       NewTorrent().WithOriginalYear(1890).WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "error - year in future",
			Actual:     NewTorrent().WithOriginalYear(3000).WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:      "valid - matches reference",
			Actual:    NewTorrent().WithTitle("Symphony").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			Reference: NewTorrent().WithTitle("Symphony").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:  true,
		},
		{
			Name:       "info - small difference from reference",
			Actual:     NewTorrent().WithOriginalYear(1963).WithEdition("Deutsche Grammophon", "DG-479-0334", 1963).Build(),
			Reference:  NewTorrent().WithOriginalYear(1965).WithEdition("Deutsche Grammophon", "DG-479-0334", 1965).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "warning - large difference from reference",
			Actual:     NewTorrent().WithOriginalYear(1963).WithEdition("Deutsche Grammophon", "DG-479-0334", 1963).Build(),
			Reference:  NewTorrent().WithOriginalYear(1970).WithEdition("Deutsche Grammophon", "DG-479-0334", 1970).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:         "warning - edition year before album year",
			Actual:       NewTorrent().WithOriginalYear(1990).WithEdition("Label", "CAT123", 1985).Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:     "valid - edition year after album year",
			Actual:   NewTorrent().WithOriginalYear(1963).WithEdition("Label", "CAT123", 2010).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - same edition and album year",
			Actual:   NewTorrent().WithOriginalYear(1963).WithEdition("Label", "CAT123", 1963).Build(),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.YearFieldUsage(tt.Actual, tt.Reference)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues {
					switch issue.Level {
					case domain.LevelError:
						errorCount++
					case domain.LevelWarning:
						warningCount++
					case domain.LevelInfo:
						infoCount++
					}
				}

				if tt.WantErrors > 0 && errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}
				if tt.WantWarnings > 0 && warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}
				if tt.WantInfo > 0 && infoCount != tt.WantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}
