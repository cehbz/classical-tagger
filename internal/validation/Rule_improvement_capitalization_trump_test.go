package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CapitalizationTrump(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name      string
		Actual    *domain.Torrent
		Reference *domain.Torrent
		WantPass  bool
		WantInfo  int
	}{
		{
			Name:     "pass - no reference",
			Actual:   NewTorrent().WithTitle("Symphony No. 5").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:      "info - no improvement",
			Actual:    buildTorrentWithBadCaps(),
			Reference: buildTorrentWithGoodCaps(),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "info - significant improvement",
			Actual:    buildTorrentWithGoodCaps(),
			Reference: buildTorrentWithBadCaps(),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "info - same quality",
			Actual:    buildTorrentWithGoodCaps(),
			Reference: buildTorrentWithGoodCaps(),
			WantPass:  false,
			WantInfo:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.CapitalizationTrump(tt.Actual, tt.Reference)

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

func TestCountCapitalizationIssues(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name      string
		Album     *domain.Torrent
		WantCount int
	}{
		{
			Name:      "good capitalization",
			Album:     buildTorrentWithGoodCaps(),
			WantCount: 0,
		},
		{
			Name:      "bad capitalization",
			Album:     buildTorrentWithBadCaps(),
			WantCount: 4, // Title + 3 track titles
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			count := rules.countCapitalizationIssues(tt.Album)

			if count != tt.WantCount {
				t.Errorf("countCapitalizationIssues() = %d, want %d", count, tt.WantCount)
			}
		})
	}
}
