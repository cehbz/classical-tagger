package tagging

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestFormatArtists tests formatting multiple artists according to classical music rules.
func TestFormatArtists(t *testing.T) {
	tests := []struct {
		Name    string
		Artists []domain.Artist
		Want    string
	}{
		{
			Name:    "single soloist",
			Artists: []domain.Artist{{Name: "Glenn Gould", Role: domain.RoleSoloist}},
			Want:    "Glenn Gould",
		},
		{
			Name: "soloist and ensemble",
			Artists: []domain.Artist{
				{Name: "Anne-Sophie Mutter", Role: domain.RoleSoloist},
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
			Want: "Anne-Sophie Mutter, Berlin Philharmonic",
		},
		{
			Name: "soloist, ensemble, and conductor",
			Artists: []domain.Artist{
				{Name: "Yo-Yo Ma", Role: domain.RoleSoloist},
				{Name: "Chicago Symphony Orchestra", Role: domain.RoleEnsemble},
				{Name: "Daniel Barenboim", Role: domain.RoleConductor},
			},
			Want: "Yo-Yo Ma, Chicago Symphony Orchestra, Daniel Barenboim",
		},
		{
			Name: "multiple soloists",
			Artists: []domain.Artist{
				{Name: "Martha Argerich", Role: domain.RoleSoloist},
				{Name: "Daniel Barenboim", Role: domain.RoleSoloist},
			},
			Want: "Martha Argerich, Daniel Barenboim",
		},
		{
			Name: "just ensemble",
			Artists: []domain.Artist{
				{Name: "The Academy of Ancient Music", Role: domain.RoleEnsemble},
			},
			Want: "The Academy of Ancient Music",
		},
		{
			Name: "ensemble and conductor",
			Artists: []domain.Artist{
				{Name: "London Symphony Orchestra", Role: domain.RoleEnsemble},
				{Name: "Claudio Abbado", Role: domain.RoleConductor},
			},
			Want: "London Symphony Orchestra, Claudio Abbado",
		},
		{
			Name:    "empty artists",
			Artists: []domain.Artist{},
			Want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := FormatArtists(tt.Artists)
			if got != tt.Want {
				t.Errorf("FormatArtists() = %q, want %q", got, tt.Want)
			}
		})
	}
}

// TestDetermineAlbumArtist tests determining the album artist from an album's tracks.
func TestDetermineAlbumArtist(t *testing.T) {
	tests := []struct {
		Name               string
		Album              *domain.Album
		WantArtist         string
		WantUniversalCount int
	}{
		{
			Name: "single performer across all tracks",
			Album: func() *domain.Album {
				composer := domain.Artist{Name: "Bach", Role: domain.RoleComposer}
				performer := domain.Artist{Name: "Glenn Gould", Role: domain.RoleSoloist}
				album := &domain.Album{
					Title: "Goldberg Variations", OriginalYear: 1981,
					Tracks: []*domain.Track{
						&domain.Track{Disc: 1, Track: 1, Title: "Aria", Artists: []domain.Artist{composer, performer}},
						&domain.Track{Disc: 1, Track: 2, Title: "Variation 1", Artists: []domain.Artist{composer, performer}},
						&domain.Track{Disc: 1, Track: 3, Title: "Variation 2", Artists: []domain.Artist{composer, performer}},
					},
				}

				return album
			}(),
			WantArtist:         "Glenn Gould",
			WantUniversalCount: 1,
		},
		{
			Name: "single ensemble across all tracks",
			Album: func() *domain.Album {
				composer := domain.Artist{Name: "Mozart", Role: domain.RoleComposer}
				ensemble := domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}
				conductor := domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}

				album := &domain.Album{Title: "Symphony No. 40", OriginalYear: 1975,
					Tracks: []*domain.Track{
						&domain.Track{Disc: 1, Track: 1, Title: "I. Allegro", Artists: []domain.Artist{composer, ensemble, conductor}},
						&domain.Track{Disc: 1, Track: 2, Title: "II. Andante", Artists: []domain.Artist{composer, ensemble, conductor}},
					},
				}

				return album
			}(),
			WantArtist:         "Vienna Philharmonic, Herbert von Karajan",
			WantUniversalCount: 2,
		},
		{
			Name: "varying performers - returns empty",
			Album: func() *domain.Album {
				composer := domain.Artist{Name: "Various", Role: domain.RoleComposer}
				performer1 := domain.Artist{Name: "Artist 1", Role: domain.RoleSoloist}
				performer2 := domain.Artist{Name: "Artist 2", Role: domain.RoleSoloist}

				album := &domain.Album{Title: "Compilation", OriginalYear: 2020,
					Tracks: []*domain.Track{
						&domain.Track{Disc: 1, Track: 1, Title: "Track 1", Artists: []domain.Artist{composer, performer1}},
						&domain.Track{Disc: 1, Track: 2, Title: "Track 2", Artists: []domain.Artist{composer, performer2}},
					},
				}

				return album
			}(),
			WantArtist:         "",
			WantUniversalCount: 0,
		},
		{
			Name: "empty album",
			Album: &domain.Album{Title: "Empty", OriginalYear: 2020},
			WantArtist:         "",
			WantUniversalCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gotArtist, gotUniversal := DetermineAlbumArtist(tt.Album)
			if gotArtist != tt.WantArtist {
				t.Errorf("DetermineAlbumArtist() artist = %q, want %q", gotArtist, tt.WantArtist)
			}
			if len(gotUniversal) != tt.WantUniversalCount {
				t.Errorf("DetermineAlbumArtist() universal count = %d, want %d", len(gotUniversal), tt.WantUniversalCount)
			}
		})
	}
}
