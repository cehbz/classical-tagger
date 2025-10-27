package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_MultiDiscFolderSorting(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Album
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "pass - single disc",
			Actual:   buildAlbumWithFilenames("01 - Track.flac"),
			WantPass: true,
		},
		{
			Name:     "pass - no folders",
			Actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {2, 1}}),
			WantPass: true,
		},
		{
			Name: "pass - properly padded folders (10+ discs)",
			Actual: buildAlbumWithFilenames(
				"CD01/01 - Track.flac",
				"CD02/01 - Track.flac",
				"CD10/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "info - unpadded folders with 10+ discs",
			Actual: buildAlbumWithFilenamesAndDiscs(
				[]string{"CD1/01 - Track.flac", "CD2/01 - Track.flac", "CD10/01 - Track.flac"},
				[]int{1, 2, 10},
			),
			WantPass: false,
			WantInfo: 2, // CD1 and CD2 both need padding
		},
		{
			Name: "pass - two digit folders",
			Actual: buildAlbumWithFilenames(
				"CD1/01 - Track.flac",
				"CD2/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "pass - Disc naming with padding",
			Actual: buildAlbumWithFilenames(
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

// buildAlbumWithFilenamesAndDiscs creates album with specific filenames and disc numbers
func buildAlbumWithFilenamesAndDiscs(filenames []string, discs []int) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	artists := []domain.Artist{composer, ensemble}

	tracks := make([]*domain.Track, len(filenames))
	for i, filename := range filenames {
		disc := discs[i]
		tracks[i] = &domain.Track{Disc: disc, Track: i + 1, Title: "Track", Artists: artists, Name: filename}
	}
	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: tracks}
}
