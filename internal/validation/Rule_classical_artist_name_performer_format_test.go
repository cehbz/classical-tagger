package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

type artistSpec struct {
	Name string
	Role domain.Role
}

func TestRules_PerformerFormat(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Torrent
		WantPass   bool
		WantErrors int
		WantInfo   int
	}{
		{
			Name:     "valid - soloist, ensemble, conductor",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Maurizio Pollini", Role: domain.RoleSoloist}, domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Claudio Abbado", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true, // Info messages don't fail
			WantInfo: 1,
		},
		{
			Name:     "valid - just ensemble and conductor",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name:     "valid - just ensemble (no conductor)",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Emerson String Quartet", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name:     "valid - multiple soloists",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Mozart", Role: domain.RoleComposer}, domain.Artist{Name: "Anne-Sophie Mutter", Role: domain.RoleSoloist}, domain.Artist{Name: "Yo-Yo Ma", Role: domain.RoleSoloist}, domain.Artist{Name: "Chamber Orchestra of Europe", Role: domain.RoleEnsemble}, domain.Artist{Name: "Daniel Barenboim", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name:       "invalid - only composer (no performers)",
			Actual:     NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}).Build().Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "invalid - only composer and arranger",
			Actual:     NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Busoni", Role: domain.RoleArranger}).Build().Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - soloist only (solo piano)",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Chopin", Role: domain.RoleComposer}, domain.Artist{Name: "Martha Argerich", Role: domain.RoleSoloist}).Build().Build(),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name:     "valid - opera with multiple soloists and ensemble",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Verdi", Role: domain.RoleComposer}, domain.Artist{Name: "Renée Fleming", Role: domain.RoleSoloist}, domain.Artist{Name: "Plácido Domingo", Role: domain.RoleSoloist}, domain.Artist{Name: "Bryn Terfel", Role: domain.RoleSoloist}, domain.Artist{Name: "Metropolitan Opera Orchestra", Role: domain.RoleEnsemble}, domain.Artist{Name: "James Levine", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name: "valid - no info",
			Actual: &domain.Torrent{
				Title:        "Classical Album",
				OriginalYear: 1963,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Symphony No. 5: Itzhak Perlman, Berlin Philharmonic, Claudio Abbado",
						Artists: []domain.Artist{
							{Name: "Beethoven", Role: domain.RoleComposer},
							{Name: "Itzhak Perlman", Role: domain.RoleSoloist},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
							{Name: "Claudio Abbado", Role: domain.RoleConductor},
						},
					},
				},
			},
			WantPass: true,
			WantInfo: 0,
		},
	}

	for _, tt := range tests {
		for _, track := range tt.Actual.Tracks() {
			t.Run(tt.Name, func(t *testing.T) {
				result := rules.PerformerFormat(track, nil, nil, nil)

				if result.Passed() != tt.WantPass {
					t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
				}

				if !tt.WantPass {
					errorCount := 0
					infoCount := 0
					for _, issue := range result.Issues {
						if issue.Level == domain.LevelError {
							errorCount++
						} else if issue.Level == domain.LevelInfo {
							infoCount++
						}
					}

					if errorCount != tt.WantErrors {
						t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
					}

					for _, issue := range result.Issues {
						t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
					}
				}
			})
		}
	}
}

func TestFormatArtistsByRole(t *testing.T) {
	tests := []struct {
		Name       string
		Soloists   []domain.Artist
		Ensembles  []domain.Artist
		Conductors []domain.Artist
		Want       string
	}{
		{
			Name: "all roles present",
			Soloists: []domain.Artist{
				domain.Artist{Name: "Maurizio Pollini", Role: domain.RoleSoloist},
			},
			Ensembles: []domain.Artist{
				domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
			Conductors: []domain.Artist{
				domain.Artist{Name: "Claudio Abbado", Role: domain.RoleConductor},
			},
			Want: "Maurizio Pollini, Berlin Philharmonic, Claudio Abbado",
		},
		{
			Name: "multiple soloists",
			Soloists: []domain.Artist{
				domain.Artist{Name: "Anne-Sophie Mutter", Role: domain.RoleSoloist},
				domain.Artist{Name: "Yo-Yo Ma", Role: domain.RoleSoloist},
			},
			Ensembles: []domain.Artist{
				domain.Artist{Name: "Chamber Orchestra", Role: domain.RoleEnsemble},
			},
			Conductors: []domain.Artist{
				domain.Artist{Name: "Daniel Barenboim", Role: domain.RoleConductor},
			},
			Want: "Anne-Sophie Mutter, Yo-Yo Ma, Chamber Orchestra, Daniel Barenboim",
		},
		{
			Name: "just ensemble and conductor",
			Ensembles: []domain.Artist{
				domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},
			},
			Conductors: []domain.Artist{
				domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			},
			Want: "Vienna Philharmonic, Herbert von Karajan",
		},
		{
			Name: "just ensemble",
			Ensembles: []domain.Artist{
				domain.Artist{Name: "Emerson String Quartet", Role: domain.RoleEnsemble},
			},
			Want: "Emerson String Quartet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := formatArtistsByRole(tt.Soloists, tt.Ensembles, tt.Conductors)
			if got != tt.Want {
				t.Errorf("formatArtistsByRole() = %q, want %q", got, tt.Want)
			}
		})
	}
}
