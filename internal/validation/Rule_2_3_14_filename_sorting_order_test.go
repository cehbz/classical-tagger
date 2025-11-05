package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FilenameSortingOrder(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Torrent
		WantPass   bool
		WantIssues int
	}{
		{
			Name: "valid - correct sorting with zero padding",
			Actual: buildTorrentWithFilenames(
				"01 - First.flac",
				"02 - Second.flac",
				"03 - Third.flac",
			),
			WantPass: true,
		},
		{
			Name: "invalid - lexicographic sort without zero padding is incorrect",
			Actual: buildTorrentWithFilenames(
				"1 - First.flac",
				"2 - Second.flac",
				"3 - Third.flac",
				"10 - Tenth.flac",
			),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "invalid - no zero padding causes sorting issue",
			Actual: buildTorrentWithTrackFilenames(
				trackFile{1, "1 - First.flac"},
				trackFile{2, "2 - Second.flac"},
				trackFile{10, "10 - Tenth.flac"},
				trackFile{3, "3 - Third.flac"}, // Will sort before "10"
			),
			WantPass:   false,
			WantIssues: 1, // Track 10 sorts before track 3
		},
		{
			Name: "invalid - incorrect ordering",
			Actual: buildTorrentWithTrackFilenames(
				trackFile{1, "02 - Second.flac"}, // Wrong filename for track 1
				trackFile{2, "01 - First.flac"},  // Wrong filename for track 2
			),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "valid - single track",
			Actual: buildTorrentWithFilenames(
				"Complete Work.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid - multi-disc with proper organization",
			Actual: buildMultiDiscTorrentWithFilenames(
				[]trackFile{
					{1, "CD1/01 - First.flac"},
					{2, "CD1/02 - Second.flac"},
				},
				[]trackFile{
					{1, "CD2/01 - Third.flac"},
					{2, "CD2/02 - Fourth.flac"},
				},
			),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.FilenameSortingOrder(tt.Actual, nil)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass && len(result.Issues) != tt.WantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues), tt.WantIssues)
				for _, issue := range result.Issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}
