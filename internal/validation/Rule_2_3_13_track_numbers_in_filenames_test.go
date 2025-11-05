package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_TrackNumbersInFilenames(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Torrent
		WantPass   bool
		WantIssues int
	}{
		{
			Name:     "valid track numbers with dash",
			Actual:   NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("01 - Symphony No. 5.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("02 - Concerto.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(3).WithTitle("Symphony No. 5").WithFilename("03 - Finale.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid track numbers with underscore",
			Actual:   NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("01_Symphony No. 5.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("02_Concerto.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid track numbers no padding",
			Actual:   NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("1 - Symphony No. 5.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("2 - Concerto.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(3).WithTitle("Symphony No. 5").WithFilename("10 - Finale.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid track numbers with period",
			Actual:   NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("01. Symphony No. 5.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("02. Concerto.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
		},
		{
			Name:       "missing track numbers",
			Actual:     NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("Symphony No. 5.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("Concerto.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name:       "some missing track numbers",
			Actual:     NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("01 - Symphony No. 5.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("Concerto.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(3).WithTitle("Symphony No. 5").WithFilename("03 - Finale.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:     "single track exception",
			Actual:   NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("Complete Work.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true, // Exception: single tracks don't need numbers
		},
		{
			Name:     "multi-disc with subfolder",
			Actual:   NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("CD1/01 - First Movement.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("CD1/02 - Second Movement.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(3).WithTitle("Symphony No. 5").WithFilename("CD2/01 - Third Movement.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
		},
		{
			Name:       "multi-disc missing numbers in subfolder",
			Actual:     NewTorrent().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTrack(1).WithTitle("Symphony No. 5").WithFilename("CD1/01 - First Movement.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().AddTrack().WithTrack(2).WithTitle("Symphony No. 5").WithFilename("CD2/Second Movement.flac").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass:   false,
			WantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.TrackNumbersInFilenames(tt.Actual, nil)

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
