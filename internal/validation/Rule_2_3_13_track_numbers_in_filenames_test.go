package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TrackNumbersInFilenames(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name: "valid track numbers with dash",
			Actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5.flac",
				"02 - Concerto.flac",
				"03 - Finale.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid track numbers with underscore",
			Actual: buildAlbumWithFilenames(
				"01_Symphony No. 5.flac",
				"02_Concerto.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid track numbers no padding",
			Actual: buildAlbumWithFilenames(
				"1 - Symphony No. 5.flac",
				"2 - Concerto.flac",
				"10 - Finale.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid track numbers with period",
			Actual: buildAlbumWithFilenames(
				"01. Symphony No. 5.flac",
				"02. Concerto.flac",
			),
			WantPass: true,
		},
		{
			Name: "missing track numbers",
			Actual: buildAlbumWithFilenames(
				"Symphony No. 5.flac",
				"Concerto.flac",
			),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name: "some missing track numbers",
			Actual: buildAlbumWithFilenames(
				"01 - Symphony No. 5.flac",
				"Concerto.flac",
				"03 - Finale.flac",
			),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "single track exception",
			Actual: buildAlbumWithFilenames(
				"Complete Work.flac",
			),
			WantPass: true, // Exception: single tracks don't need numbers
		},
		{
			Name: "multi-disc with subfolder",
			Actual: buildAlbumWithFilenames(
				"CD1/01 - First Movement.flac",
				"CD1/02 - Second Movement.flac",
				"CD2/01 - Third Movement.flac",
			),
			WantPass: true,
		},
		{
			Name: "multi-disc missing numbers in subfolder",
			Actual: buildAlbumWithFilenames(
				"CD1/01 - First Movement.flac",
				"CD2/Second Movement.flac",
			),
			WantPass:   false,
			WantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Use actual for both parameters (reference not used by this rule)
			result := rules.TrackNumbersInFilenames(tt.Actual, tt.Actual)

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

// Helper function to build an album with specific filenames
func buildAlbumWithFilenames(filenames ...string) *domain.Album {
	tracks := make([]*domain.Track, len(filenames))
	for i, _ := range filenames {
		tracks[i] = &domain.Track{
			Disc:  1,
			Track: i + 1,
			Title: "Symphony No. 5",
			Artists: []domain.Artist{
				domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer},
				domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},
				domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			},
		}
	}
	return &domain.Album{
		Title:        "Beethoven Symphonies",
		OriginalYear: 1963,
		Tracks:       tracks,
	}
}
