package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_MultiDiscFolderSorting(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name     string
		actual   *domain.Album
		wantPass bool
		wantInfo int
	}{
		{
			name:     "pass - single disc",
			actual:   buildAlbumWithFilenames("01 - Track.flac"),
			wantPass: true,
		},
		{
			name:     "pass - no folders",
			actual:   buildAlbumWithDiscTracks([]discTrack{{1, 1}, {2, 1}}),
			wantPass: true,
		},
		{
			name: "pass - properly padded folders (10+ discs)",
			actual: buildAlbumWithFilenames(
				"CD01/01 - Track.flac",
				"CD02/01 - Track.flac",
				"CD10/01 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "info - unpadded folders with 10+ discs",
			actual: buildAlbumWithFilenamesAndDiscs(
				[]string{"CD1/01 - Track.flac", "CD2/01 - Track.flac", "CD10/01 - Track.flac"},
				[]int{1, 2, 10},
			),
			wantPass: false,
			wantInfo: 2, // CD1 and CD2 both need padding
		},
		{
			name: "pass - two digit folders",
			actual: buildAlbumWithFilenames(
				"CD1/01 - Track.flac",
				"CD2/01 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "pass - Disc naming with padding",
			actual: buildAlbumWithFilenames(
				"Disc01/01 - Track.flac",
				"Disc02/01 - Track.flac",
				"Disc10/01 - Track.flac",
			),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.MultiDiscFolderSorting(tt.actual, tt.actual)

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

// buildAlbumWithFilenamesAndDiscs creates album with specific filenames and disc numbers
func buildAlbumWithFilenamesAndDiscs(filenames []string, discs []int) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	artists := []domain.Artist{composer, ensemble}

	album, _ := domain.NewAlbum("Album", 1963)
	for i, filename := range filenames {
		disc := discs[i]
		track, _ := domain.NewTrack(disc, i+1, "Track", artists)
		track = track.WithName(filename)
		album.AddTrack(track)
	}

	return album
}
