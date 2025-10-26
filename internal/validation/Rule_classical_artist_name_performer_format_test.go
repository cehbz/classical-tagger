package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_PerformerFormat(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantErrors int
		wantInfo   int
	}{
		{
			name: "valid - soloist, ensemble, conductor",
			actual: buildAlbumWithArtists(
				"Ludwig van Beethoven", domain.RoleComposer,
				"Maurizio Pollini", domain.RoleSoloist,
				"Berlin Philharmonic", domain.RoleEnsemble,
				"Claudio Abbado", domain.RoleConductor,
			),
			wantPass: true, // Info messages don't fail
			wantInfo: 1,
		},
		{
			name: "valid - just ensemble and conductor",
			actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Vienna Philharmonic", domain.RoleEnsemble,
				"Herbert von Karajan", domain.RoleConductor,
			),
			wantPass: true,
			wantInfo: 1,
		},
		{
			name: "valid - just ensemble (no conductor)",
			actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Emerson String Quartet", domain.RoleEnsemble,
			),
			wantPass: true,
			wantInfo: 1,
		},
		{
			name: "valid - multiple soloists",
			actual: buildAlbumWithArtists(
				"Mozart", domain.RoleComposer,
				"Anne-Sophie Mutter", domain.RoleSoloist,
				"Yo-Yo Ma", domain.RoleSoloist,
				"Chamber Orchestra of Europe", domain.RoleEnsemble,
				"Daniel Barenboim", domain.RoleConductor,
			),
			wantPass: true,
			wantInfo: 1,
		},
		{
			name: "invalid - only composer (no performers)",
			actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
			),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "invalid - only composer and arranger",
			actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Busoni", domain.RoleArranger,
			),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "valid - soloist only (solo piano)",
			actual: buildAlbumWithArtists(
				"Chopin", domain.RoleComposer,
				"Martha Argerich", domain.RoleSoloist,
			),
			wantPass: true,
			wantInfo: 1,
		},
		{
			name: "valid - opera with multiple soloists and ensemble",
			actual: buildAlbumWithArtists(
				"Verdi", domain.RoleComposer,
				"Renée Fleming", domain.RoleSoloist,
				"Plácido Domingo", domain.RoleSoloist,
				"Bryn Terfel", domain.RoleSoloist,
				"Metropolitan Opera Orchestra", domain.RoleEnsemble,
				"James Levine", domain.RoleConductor,
			),
			wantPass: true,
			wantInfo: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.PerformerFormat(tt.actual, tt.actual)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass {
				errorCount := 0
				infoCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelError {
						errorCount++
					} else if issue.Level() == domain.LevelInfo {
						infoCount++
					}
				}
				
				if errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}
				
				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

func TestFormatArtistsByRole(t *testing.T) {
	tests := []struct {
		name       string
		soloists   []domain.Artist
		ensembles  []domain.Artist
		conductors []domain.Artist
		want       string
	}{
		{
			name: "all roles present",
			soloists: []domain.Artist{
				mustCreateArtist("Maurizio Pollini", domain.RoleSoloist),
			},
			ensembles: []domain.Artist{
				mustCreateArtist("Berlin Philharmonic", domain.RoleEnsemble),
			},
			conductors: []domain.Artist{
				mustCreateArtist("Claudio Abbado", domain.RoleConductor),
			},
			want: "Maurizio Pollini, Berlin Philharmonic, Claudio Abbado",
		},
		{
			name: "multiple soloists",
			soloists: []domain.Artist{
				mustCreateArtist("Anne-Sophie Mutter", domain.RoleSoloist),
				mustCreateArtist("Yo-Yo Ma", domain.RoleSoloist),
			},
			ensembles: []domain.Artist{
				mustCreateArtist("Chamber Orchestra", domain.RoleEnsemble),
			},
			conductors: []domain.Artist{
				mustCreateArtist("Daniel Barenboim", domain.RoleConductor),
			},
			want: "Anne-Sophie Mutter, Yo-Yo Ma, Chamber Orchestra, Daniel Barenboim",
		},
		{
			name: "just ensemble and conductor",
			ensembles: []domain.Artist{
				mustCreateArtist("Vienna Philharmonic", domain.RoleEnsemble),
			},
			conductors: []domain.Artist{
				mustCreateArtist("Herbert von Karajan", domain.RoleConductor),
			},
			want: "Vienna Philharmonic, Herbert von Karajan",
		},
		{
			name: "just ensemble",
			ensembles: []domain.Artist{
				mustCreateArtist("Emerson String Quartet", domain.RoleEnsemble),
			},
			want: "Emerson String Quartet",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatArtistsByRole(tt.soloists, tt.ensembles, tt.conductors)
			if got != tt.want {
				t.Errorf("formatArtistsByRole() = %q, want %q", got, tt.want)
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
		artist, _ := domain.NewArtist(name, role)
		artists = append(artists, artist)
	}
	
	track, _ := domain.NewTrack(1, 1, "Symphony No. 5", artists)
	album, _ := domain.NewAlbum("Classical Album", 1963)
	album.AddTrack(track)
	return album
}

// Helper to create artist without error handling
func mustCreateArtist(name string, role domain.Role) domain.Artist {
	artist, err := domain.NewArtist(name, role)
	if err != nil {
		panic(err)
	}
	return artist
}
