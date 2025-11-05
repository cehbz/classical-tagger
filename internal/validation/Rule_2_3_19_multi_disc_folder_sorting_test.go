package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_MultiDiscFolderSorting(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual *domain.Torrent
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "pass - single disc",
			Actual:   buildTorrentWithFilenames("01 - Track.flac"),
			WantPass: true,
		},
		{
			Name:     "pass - no folders",
			Actual:   buildTorrentWithDiscTracks([]discTrack{{1, 1}, {2, 1}}),
			WantPass: true,
		},
		{
			Name: "pass - properly padded folders (10+ discs)",
			Actual: buildTorrentWithFilenames(
				"CD01/01 - Track.flac",
				"CD02/01 - Track.flac",
				"CD10/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "info - unpadded folders with 10+ discs",
			Actual: buildTorrentWithFilenamesAndDiscs(
				[]string{"CD1/01 - Track.flac", "CD2/01 - Track.flac", "CD10/01 - Track.flac"},
				[]int{1, 2, 10},
			),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name: "pass - two digit folders",
			Actual: buildTorrentWithFilenames(
				"CD1/01 - Track.flac",
				"CD2/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "pass - Disc naming with padding",
			Actual: buildTorrentWithFilenames(
				"Disc01/01 - Track.flac",
				"Disc02/01 - Track.flac",
				"Disc10/01 - Track.flac",
			),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.MultiDiscFolderSorting(tt.Actual, tt.Actual)

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
