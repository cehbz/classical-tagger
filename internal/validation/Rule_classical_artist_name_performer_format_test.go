package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_PerformerFormat(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantErrors int
		WantInfo   int
	}{
		{
			Name: "valid - soloist, ensemble, conductor",
			Actual: buildAlbumWithArtists(
				"Ludwig van Beethoven", domain.RoleComposer,
				"Maurizio Pollini", domain.RoleSoloist,
				"Berlin Philharmonic", domain.RoleEnsemble,
				"Claudio Abbado", domain.RoleConductor,
			),
			WantPass: true, // Info messages don't fail
			WantInfo: 1,
		},
		{
			Name: "valid - just ensemble and conductor",
			Actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Vienna Philharmonic", domain.RoleEnsemble,
				"Herbert von Karajan", domain.RoleConductor,
			),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name: "valid - just ensemble (no conductor)",
			Actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Emerson String Quartet", domain.RoleEnsemble,
			),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name: "valid - multiple soloists",
			Actual: buildAlbumWithArtists(
				"Mozart", domain.RoleComposer,
				"Anne-Sophie Mutter", domain.RoleSoloist,
				"Yo-Yo Ma", domain.RoleSoloist,
				"Chamber Orchestra of Europe", domain.RoleEnsemble,
				"Daniel Barenboim", domain.RoleConductor,
			),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name: "invalid - only composer (no performers)",
			Actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
			),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "invalid - only composer and arranger",
			Actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Busoni", domain.RoleArranger,
			),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "valid - soloist only (solo piano)",
			Actual: buildAlbumWithArtists(
				"Chopin", domain.RoleComposer,
				"Martha Argerich", domain.RoleSoloist,
			),
			WantPass: true,
			WantInfo: 1,
		},
		{
			Name: "valid - opera with multiple soloists and ensemble",
			Actual: buildAlbumWithArtists(
				"Verdi", domain.RoleComposer,
				"Renée Fleming", domain.RoleSoloist,
				"Plácido Domingo", domain.RoleSoloist,
				"Bryn Terfel", domain.RoleSoloist,
				"Metropolitan Opera Orchestra", domain.RoleEnsemble,
				"James Levine", domain.RoleConductor,
			),
			WantPass: true,
			WantInfo: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.PerformerFormat(tt.Actual, tt.Actual)

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
				mustCreateArtist("Maurizio Pollini", domain.RoleSoloist),
			},
			Ensembles: []domain.Artist{
				mustCreateArtist("Berlin Philharmonic", domain.RoleEnsemble),
			},
			Conductors: []domain.Artist{
				mustCreateArtist("Claudio Abbado", domain.RoleConductor),
			},
			Want: "Maurizio Pollini, Berlin Philharmonic, Claudio Abbado",
		},
		{
			Name: "multiple soloists",
			Soloists: []domain.Artist{
				mustCreateArtist("Anne-Sophie Mutter", domain.RoleSoloist),
				mustCreateArtist("Yo-Yo Ma", domain.RoleSoloist),
			},
			Ensembles: []domain.Artist{
				mustCreateArtist("Chamber Orchestra", domain.RoleEnsemble),
			},
			Conductors: []domain.Artist{
				mustCreateArtist("Daniel Barenboim", domain.RoleConductor),
			},
			Want: "Anne-Sophie Mutter, Yo-Yo Ma, Chamber Orchestra, Daniel Barenboim",
		},
		{
			Name: "just ensemble and conductor",
			Ensembles: []domain.Artist{
				mustCreateArtist("Vienna Philharmonic", domain.RoleEnsemble),
			},
			Conductors: []domain.Artist{
				mustCreateArtist("Herbert von Karajan", domain.RoleConductor),
			},
			Want: "Vienna Philharmonic, Herbert von Karajan",
		},
		{
			Name: "just ensemble",
			Ensembles: []domain.Artist{
				mustCreateArtist("Emerson String Quartet", domain.RoleEnsemble),
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

// Helper to build album with specific artists
func buildAlbumWithArtists(artistSpecs ...interface{}) *domain.Album {
	if len(artistSpecs)%2 != 0 {
		panic("artistSpecs must be pairs of (name, role)")
	}

	var artists []domain.Artist
	for i := 0; i < len(artistSpecs); i += 2 {
		name := artistSpecs[i].(string)
		role := artistSpecs[i+1].(domain.Role)
		artist := domain.Artist{Name: name, Role: role}
		artists = append(artists, artist)
	}

	track := domain.Track{Disc: 1, Track: 1, Title: "Symphony No. 5", Artists: artists}
	return &domain.Album{Title: "Classical Album", OriginalYear: 1963, Tracks: []*domain.Track{&track}}
}

// Helper to create artist without error handling
func mustCreateArtist(name string, role domain.Role) domain.Artist {
	return domain.Artist{Name: name, Role: role}
}
