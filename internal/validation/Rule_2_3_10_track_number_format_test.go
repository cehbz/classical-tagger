package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TrackNumberFormat(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name     string
		actual   *domain.Album
		wantPass bool
		wantInfo int
	}{
		{
			name:     "pass - sequential numbering",
			actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {1, 3}}),
			wantPass: true,
		},
		{
			name:     "info - gap in numbering",
			actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {1, 4}}),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "pass - multi-disc starting at 1",
			actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {2, 1}, {2, 2}}),
			wantPass: true,
		},
		{
			name:     "info - disc doesn't start at 1",
			actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 2}, {2, 2}, {2, 3}}),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - multiple gaps",
			actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {1, 3}, {1, 5}}),
			wantPass: false,
			wantInfo: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.TrackNumberFormat(tt.actual, tt.actual)
			
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
