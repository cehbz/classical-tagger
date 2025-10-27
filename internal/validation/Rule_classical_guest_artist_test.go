package validation

import (
	"fmt"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_GuestArtistIdentification(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Album
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "pass - single track",
			Actual:   buildAlbumWithArtists("Beethoven", domain.RoleComposer, "Pollini", domain.RoleSoloist),
			WantPass: true,
		},
		{
			Name:     "pass - all tracks have same soloist",
			Actual:   buildAlbumWithConsistentSoloist("Pollini", 5),
			WantPass: true,
		},
		{
			Name:     "info - infrequent soloist",
			Actual:   buildAlbumWithGuestSoloist("Pollini", "Guest Artist", 5),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "pass - guest indicated in title",
			Actual:   buildAlbumWithGuestInTitle("Pollini", "Guest Artist", 5),
			WantPass: true,
		},
		{
			Name:     "pass - too few tracks to determine",
			Actual:   buildAlbumWithConsistentSoloist("Pollini", 2),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.GuestArtistIdentification(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				infoCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelInfo {
						infoCount++
					}
				}

				if infoCount != tt.WantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

// buildAlbumWithConsistentSoloist creates album with same soloist on all tracks
func buildAlbumWithConsistentSoloist(soloistName string, trackCount int) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	soloist := domain.Artist{Name: soloistName, Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

	tracks := make([]*domain.Track, trackCount)
	for i := 0; i < trackCount; i++ {
		tracks[i] = &domain.Track{
			Disc:    1,
			Track:   i + 1,
			Title:   fmt.Sprintf("Concerto No. %d", i+1),
			Artists: []domain.Artist{composer, soloist, ensemble},
		}
	}

	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: tracks}
}

// buildAlbumWithGuestSoloist creates album where one soloist appears infrequently
func buildAlbumWithGuestSoloist(mainSoloist, guestSoloist string, trackCount int) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	main := domain.Artist{Name: mainSoloist, Role: domain.RoleSoloist}
	guest := domain.Artist{Name: guestSoloist, Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

	tracks := make([]*domain.Track, trackCount)
	for i := 0; i < trackCount; i++ {
		var artists []domain.Artist
		// Guest appears only on first track
		if i == 0 {
			artists = []domain.Artist{composer, guest, ensemble}
		} else {
			artists = []domain.Artist{composer, main, ensemble}
		}

		tracks[i] = &domain.Track{Disc: 1, Track: i + 1, Title: fmt.Sprintf("Concerto No. %d", i+1), Artists: artists}
	}

	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: tracks}
}

// buildAlbumWithGuestInTitle creates album with guest indicated in title
func buildAlbumWithGuestInTitle(mainSoloist, guestSoloist string, trackCount int) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	main := domain.Artist{Name: mainSoloist, Role: domain.RoleSoloist}
	guest := domain.Artist{Name: guestSoloist, Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

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

		tracks[i] = &domain.Track{Disc: 1, Track: i + 1, Title: title, Artists: artists}
	}

	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: tracks}
}
