package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CatalogInfoInComments(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Torrent
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "pass - complete edition info",
			Actual:   buildTorrentWithCompleteEdition(),
			WantPass: true,
		},
		{
			Name:     "info - no edition info",
			Actual:   NewTorrent().WithTitle("Symphony").WithoutEdition().Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - missing label",
			Actual:   NewTorrent().WithEdition("", "CAT123", 1990).Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - missing catalog",
			Actual:   NewTorrent().WithEdition("Deutsche Grammophon", "", 1990).Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - missing year",
			Actual:   NewTorrent().WithEdition("Deutsche Grammophon", "CAT123", 0).Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - multiple missing fields",
			Actual:   NewTorrent().WithEdition("", "", 0).Build(),
			WantPass: false,
			WantInfo: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.CatalogInfoInComments(tt.Actual, nil)

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
