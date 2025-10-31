package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_GuestArtistIdentification(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Album
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "pass - single track",
			Actual:   NewAlbum().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "pass - all tracks have same soloist",
			Actual:   buildAlbumWithConsistentSoloist("Pollini", 5),
			WantPass: true,
		},
		{
			Name:     "info - infrequent soloist",
			Actual:   buildAlbumWithGuestSoloist("Pollini", "Guest Artist", 5),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "pass - guest indicated in title",
			Actual:   buildAlbumWithGuestInTitle("Pollini", "Guest Artist", 5),
			WantPass: true,
		},
		{
			Name:     "pass - too few tracks to determine",
			Actual:   buildAlbumWithConsistentSoloist("Pollini", 2),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.GuestArtistIdentification(tt.Actual, tt.Actual)

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
