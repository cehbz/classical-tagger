package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_MultiDiscTrackNumbering(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual *domain.Torrent
		WantPass     bool
		WantErrors   int
		WantWarnings int
	}{
		{
			Name: "valid - single disc",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {1, 3}},
			),
			WantPass: true,
		},
		{
			Name: "valid - two discs, both start at 1",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {2, 1}, {2, 2}},
			),
			WantPass: true,
		},
		{
			Name: "valid - three discs, all start at 1",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{
					{1, 1}, {1, 2},
					{2, 1}, {2, 2}, {2, 3},
					{3, 1}, {3, 2},
				},
			),
			WantPass: true,
		},
		{
			Name: "invalid - disc 2 doesn't start at 1",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {2, 3}, {2, 4}},
			),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "invalid - disc 1 doesn't start at 1",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{{1, 2}, {1, 3}, {2, 1}, {2, 2}},
			),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "invalid - both discs don't start at 1",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{{1, 5}, {1, 6}, {2, 10}, {2, 11}},
			),
			WantPass:   false,
			WantErrors: 2,
		},
		{
			Name: "warning - gap in track numbering",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {1, 4}}, // Missing track 3
			),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name: "invalid - missing disc in sequence",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {3, 1}, {3, 2}}, // No disc 2
			),
			WantPass:   false,
			WantErrors: 1, // Missing disc 2
		},
		{
			Name: "valid - large multi-disc set",
			Actual: buildTorrentWithDiscTracks(
				[]discTrack{
					{1, 1}, {1, 2}, {1, 3},
					{2, 1}, {2, 2},
					{3, 1}, {3, 2}, {3, 3}, {3, 4},
					{4, 1}, {4, 2},
				},
			),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.MultiDiscTrackNumbering(tt.Actual, nil)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelError {
						errorCount++
					} else if issue.Level == domain.LevelWarning {
						warningCount++
					}
				}

				if tt.WantErrors > 0 && errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}
				if tt.WantWarnings > 0 && warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

