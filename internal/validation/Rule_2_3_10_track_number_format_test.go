package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TrackNumberFormat(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Album
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "pass - sequential numbering",
			Actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {1, 3}}),
			WantPass: true,
		},
		{
			Name:     "info - gap in numbering",
			Actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {1, 4}}),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "pass - multi-disc starting at 1",
			Actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {2, 1}, {2, 2}}),
			WantPass: true,
		},
		{
			Name:     "info - disc doesn't start at 1",
			Actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {2, 2}, {2, 3}}),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - multiple gaps",
			Actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 3}, {1, 5}}),
			WantPass: false,
			WantInfo: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.TrackNumberFormat(tt.Actual, tt.Actual)

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
