package tagging

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestFormatArtists tests formatting multiple artists according to classical music rules.
func TestFormatArtists(t *testing.T) {
	tests := []struct {
		name    string
		artists []domain.Artist
		want    string
	}{
		{
			name:    "single soloist",
			artists: []domain.Artist{mustArtist("Glenn Gould", domain.RoleSoloist)},
			want:    "Glenn Gould",
		},
		{
			name: "soloist and ensemble",
			artists: []domain.Artist{
				mustArtist("Anne-Sophie Mutter", domain.RoleSoloist),
				mustArtist("Berlin Philharmonic", domain.RoleEnsemble),
			},
			want: "Anne-Sophie Mutter, Berlin Philharmonic",
		},
		{
			name: "soloist, ensemble, and conductor",
			artists: []domain.Artist{
				mustArtist("Yo-Yo Ma", domain.RoleSoloist),
				mustArtist("Chicago Symphony Orchestra", domain.RoleEnsemble),
				mustArtist("Daniel Barenboim", domain.RoleConductor),
			},
			want: "Yo-Yo Ma, Chicago Symphony Orchestra, Daniel Barenboim",
		},
		{
			name: "multiple soloists",
			artists: []domain.Artist{
				mustArtist("Martha Argerich", domain.RoleSoloist),
				mustArtist("Daniel Barenboim", domain.RoleSoloist),
			},
			want: "Martha Argerich, Daniel Barenboim",
		},
		{
			name: "just ensemble",
			artists: []domain.Artist{
				mustArtist("The Academy of Ancient Music", domain.RoleEnsemble),
			},
			want: "The Academy of Ancient Music",
		},
		{
			name: "ensemble and conductor",
			artists: []domain.Artist{
				mustArtist("London Symphony Orchestra", domain.RoleEnsemble),
				mustArtist("Claudio Abbado", domain.RoleConductor),
			},
			want: "London Symphony Orchestra, Claudio Abbado",
		},
		{
			name:    "empty artists",
			artists: []domain.Artist{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatArtists(tt.artists)
			if got != tt.want {
				t.Errorf("FormatArtists() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDetermineAlbumArtist tests determining the album artist from an album's tracks.
func TestDetermineAlbumArtist(t *testing.T) {
	tests := []struct {
		name               string
		album              *domain.Album
		wantArtist         string
		wantUniversalCount int
	}{
		{
			name: "single performer across all tracks",
			album: func() *domain.Album {
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				performer, _ := domain.NewArtist("Glenn Gould", domain.RoleSoloist)

				album, _ := domain.NewAlbum("Goldberg Variations", 1981)
				track1, _ := domain.NewTrack(1, 1, "Aria", []domain.Artist{composer, performer})
				track2, _ := domain.NewTrack(1, 2, "Variation 1", []domain.Artist{composer, performer})
				track3, _ := domain.NewTrack(1, 3, "Variation 2", []domain.Artist{composer, performer})

				album.AddTrack(track1)
				album.AddTrack(track2)
				album.AddTrack(track3)

				return album
			}(),
			wantArtist:         "Glenn Gould",
			wantUniversalCount: 1,
		},
		{
			name: "single ensemble across all tracks",
			album: func() *domain.Album {
				composer, _ := domain.NewArtist("Mozart", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Philharmonic", domain.RoleEnsemble)
				conductor, _ := domain.NewArtist("Herbert von Karajan", domain.RoleConductor)

				album, _ := domain.NewAlbum("Symphony No. 40", 1975)
				track1, _ := domain.NewTrack(1, 1, "I. Allegro", []domain.Artist{composer, ensemble, conductor})
				track2, _ := domain.NewTrack(1, 2, "II. Andante", []domain.Artist{composer, ensemble, conductor})

				album.AddTrack(track1)
				album.AddTrack(track2)

				return album
			}(),
			wantArtist:         "Vienna Philharmonic, Herbert von Karajan",
			wantUniversalCount: 2,
		},
		{
			name: "varying performers - returns empty",
			album: func() *domain.Album {
				composer, _ := domain.NewArtist("Various", domain.RoleComposer)
				performer1, _ := domain.NewArtist("Artist 1", domain.RoleSoloist)
				performer2, _ := domain.NewArtist("Artist 2", domain.RoleSoloist)

				album, _ := domain.NewAlbum("Compilation", 2020)
				track1, _ := domain.NewTrack(1, 1, "Track 1", []domain.Artist{composer, performer1})
				track2, _ := domain.NewTrack(1, 2, "Track 2", []domain.Artist{composer, performer2})

				album.AddTrack(track1)
				album.AddTrack(track2)

				return album
			}(),
			wantArtist:         "",
			wantUniversalCount: 0,
		},
		{
			name: "empty album",
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Empty", 2020)
				return album
			}(),
			wantArtist:         "",
			wantUniversalCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArtist, gotUniversal := DetermineAlbumArtist(tt.album)
			if gotArtist != tt.wantArtist {
				t.Errorf("DetermineAlbumArtist() artist = %q, want %q", gotArtist, tt.wantArtist)
			}
			if len(gotUniversal) != tt.wantUniversalCount {
				t.Errorf("DetermineAlbumArtist() universal count = %d, want %d", len(gotUniversal), tt.wantUniversalCount)
			}
		})
	}
}

// mustArtist is a test helper that creates an artist and panics on error.
func mustArtist(name string, role domain.Role) domain.Artist {
	artist, err := domain.NewArtist(name, role)
	if err != nil {
		panic(err)
	}
	return artist
}
