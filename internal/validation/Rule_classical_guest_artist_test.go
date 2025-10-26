package validation

import (
	"fmt"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_GuestArtistIdentification(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name     string
		actual   *domain.Album
		wantPass bool
		wantInfo int
	}{
		{
			name:     "pass - single track",
			actual:   buildAlbumWithArtists("Beethoven", domain.RoleComposer, "Pollini", domain.RoleSoloist),
			wantPass: true,
		},
		{
			name:     "pass - all tracks have same soloist",
			actual:   buildAlbumWithConsistentSoloist("Pollini", 5),
			wantPass: true,
		},
		{
			name:     "info - infrequent soloist",
			actual:   buildAlbumWithGuestSoloist("Pollini", "Guest Artist", 5),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "pass - guest indicated in title",
			actual:   buildAlbumWithGuestInTitle("Pollini", "Guest Artist", 5),
			wantPass: true,
		},
		{
			name:     "pass - too few tracks to determine",
			actual:   buildAlbumWithConsistentSoloist("Pollini", 2),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.GuestArtistIdentification(tt.actual, tt.actual)

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

// buildAlbumWithConsistentSoloist creates album with same soloist on all tracks
func buildAlbumWithConsistentSoloist(soloistName string, trackCount int) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	soloist, _ := domain.NewArtist(soloistName, domain.RoleSoloist)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)

	album, _ := domain.NewAlbum("Album", 1963)
	for i := 0; i < trackCount; i++ {
		track, _ := domain.NewTrack(1, i+1, fmt.Sprintf("Concerto No. %d", i+1),
			[]domain.Artist{composer, soloist, ensemble})
		album.AddTrack(track)
	}

	return album
}

// buildAlbumWithGuestSoloist creates album where one soloist appears infrequently
func buildAlbumWithGuestSoloist(mainSoloist, guestSoloist string, trackCount int) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	main, _ := domain.NewArtist(mainSoloist, domain.RoleSoloist)
	guest, _ := domain.NewArtist(guestSoloist, domain.RoleSoloist)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)

	album, _ := domain.NewAlbum("Album", 1963)
	for i := 0; i < trackCount; i++ {
		var artists []domain.Artist
		// Guest appears only on first track
		if i == 0 {
			artists = []domain.Artist{composer, guest, ensemble}
		} else {
			artists = []domain.Artist{composer, main, ensemble}
		}

		track, _ := domain.NewTrack(1, i+1, fmt.Sprintf("Concerto No. %d", i+1), artists)
		album.AddTrack(track)
	}

	return album
}

// buildAlbumWithGuestInTitle creates album with guest indicated in title
func buildAlbumWithGuestInTitle(mainSoloist, guestSoloist string, trackCount int) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	main, _ := domain.NewArtist(mainSoloist, domain.RoleSoloist)
	guest, _ := domain.NewArtist(guestSoloist, domain.RoleSoloist)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)

	tracks := make([]*domain.Track, trackCount)
	for i := 0; i < trackCount; i++ {
		var artists []domain.Artist
		var title string

		// Guest appears only on first track, indicated in title
		if i == 0 {
			artists = []domain.Artist{composer, guest, ensemble}
			title = fmt.Sprintf("Concerto No. %d (feat. %s)", i+1, guestSoloist)
		} else {
			artists = []domain.Artist{composer, main, ensemble}
			title = fmt.Sprintf("Concerto No. %d", i+1)
		}

		track, _ := domain.NewTrack(1, i+1, title, artists)
		tracks[i] = track
	}

	album, _ := domain.NewAlbum("Album", 1963)
	for _, track := range tracks {
		album.AddTrack(track)
	}
	return album
}
