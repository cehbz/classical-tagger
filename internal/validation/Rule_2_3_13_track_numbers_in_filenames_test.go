package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TrackNumbersInFilenames(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid track numbers with dash",
			actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5.flac",
				"02 - Concerto.flac",
				"03 - Finale.flac",
			),
			wantPass: true,
		},
		{
			name: "valid track numbers with underscore",
			actual: buildAlbumWithFilenames(
				"01_Symphony No. 5.flac",
				"02_Concerto.flac",
			),
			wantPass: true,
		},
		{
			name: "valid track numbers no padding",
			actual: buildAlbumWithFilenames(
				"1 - Symphony No. 5.flac",
				"2 - Concerto.flac",
				"10 - Finale.flac",
			),
			wantPass: true,
		},
		{
			name: "valid track numbers with period",
			actual: buildAlbumWithFilenames(
				"01. Symphony No. 5.flac",
				"02. Concerto.flac",
			),
			wantPass: true,
		},
		{
			name: "missing track numbers",
			actual: buildAlbumWithFilenames(
				"Symphony No. 5.flac",
				"Concerto.flac",
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "some missing track numbers",
			actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5.flac",
				"Concerto.flac",
				"03 - Finale.flac",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "single track exception",
			actual: buildAlbumWithFilenames(
				"Complete Work.flac",
			),
			wantPass: true, // Exception: single tracks don't need numbers
		},
		{
			name: "multi-disc with subfolder",
			actual: buildAlbumWithFilenames(
				"CD1/01 - First Movement.flac",
				"CD1/02 - Second Movement.flac",
				"CD2/01 - Third Movement.flac",
			),
			wantPass: true,
		},
		{
			name: "multi-disc missing numbers in subfolder",
			actual: buildAlbumWithFilenames(
				"CD1/01 - First Movement.flac",
				"CD2/Second Movement.flac",
			),
			wantPass:   false,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use actual for both parameters (reference not used by this rule)
			result := rules.TrackNumbersInFilenames(tt.actual, tt.actual)

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

// Helper function to build an album with specific filenames
func buildAlbumWithFilenames(filenames ...string) *domain.Album {
	composer, _ := domain.NewArtist("Ludwig van Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Vienna Philharmonic", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Herbert von Karajan", domain.RoleConductor)
	artists := []domain.Artist{composer, ensemble, conductor}

	album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
	for i, filename := range filenames {
		track, _ := domain.NewTrack(1, i+1, "Symphony No. 5", artists)
		track = track.WithName(filename)
		album.AddTrack(track)
	}

	return album
}
