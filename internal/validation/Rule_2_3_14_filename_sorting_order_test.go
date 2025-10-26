package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FilenameSortingOrder(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - correct sorting with zero padding",
			actual: buildAlbumWithFilenames(
				"01 - First.flac",
				"02 - Second.flac",
				"03 - Third.flac",
			),
			wantPass: true,
		},
		{
			name: "invalid - lexicographic sort without zero padding is incorrect",
			actual: buildAlbumWithFilenames(
				"1 - First.flac",
				"2 - Second.flac",
				"3 - Third.flac",
				"10 - Tenth.flac",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - no zero padding causes sorting issue",
			actual: buildAlbumWithTrackFilenames(
				trackFile{1, "1 - First.flac"},
				trackFile{2, "2 - Second.flac"},
				trackFile{10, "10 - Tenth.flac"},
				trackFile{3, "3 - Third.flac"}, // Will sort before "10"
			),
			wantPass:   false,
			wantIssues: 1, // Track 10 sorts before track 3
		},
		{
			name: "invalid - incorrect ordering",
			actual: buildAlbumWithTrackFilenames(
				trackFile{1, "02 - Second.flac"}, // Wrong filename for track 1
				trackFile{2, "01 - First.flac"},  // Wrong filename for track 2
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "valid - single track",
			actual: buildAlbumWithFilenames(
				"Complete Work.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - multi-disc with proper organization",
			actual: buildMultiDiscAlbumWithFilenames(
				[]trackFile{
					{1, "CD1/01 - First.flac"},
					{2, "CD1/02 - Second.flac"},
				},
				[]trackFile{
					{1, "CD2/01 - Third.flac"},
					{2, "CD2/02 - Fourth.flac"},
				},
			),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.FilenameSortingOrder(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass && len(result.Issues()) != tt.wantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues()), tt.wantIssues)
				for _, issue := range result.Issues() {
					t.Logf("  Issue: %s", issue.Message())
				}
			}
		})
	}
}

// trackFile pairs track number with filename
type trackFile struct {
	trackNum int
	filename string
}

// buildAlbumWithTrackFilenames creates an album with specific track/filename pairs
func buildAlbumWithTrackFilenames(trackFiles ...trackFile) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	artists := []domain.Artist{composer, ensemble}

	album, _ := domain.NewAlbum("Test Album", 1963)
	for i, tf := range trackFiles {
		track, _ := domain.NewTrack(1, tf.trackNum, "Work "+string(rune('A'+i)), artists)
		track = track.WithName(tf.filename)
		album.AddTrack(track)
	}

	return album
}

// buildMultiDiscAlbumWithFilenames creates multi-disc album
func buildMultiDiscAlbumWithFilenames(disc1, disc2 []trackFile) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	artists := []domain.Artist{composer, ensemble}

	album, _ := domain.NewAlbum("Multi-Disc Album", 1963)
	for _, tf := range disc1 {
		track, _ := domain.NewTrack(1, tf.trackNum, "Work D1-"+string(rune('A'+tf.trackNum)), artists)
		track = track.WithName(tf.filename)
		album.AddTrack(track)
	}

	for _, tf := range disc2 {
		track, _ := domain.NewTrack(2, tf.trackNum, "Work D2-"+string(rune('A'+tf.trackNum)), artists)
		track = track.WithName(tf.filename)
		album.AddTrack(track)
	}
	return album
}
