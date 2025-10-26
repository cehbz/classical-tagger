package validation

import (
	"fmt"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_AlbumArtistTag(t *testing.T) {
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
			name:     "info - dominant performer",
			actual:   buildMultiTrackAlbumWithSamePerformer("Berlin Philharmonic", 5),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "info - various composers",
			actual:   buildVariousComposersAlbum(5),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "pass - no dominant performer",
			actual:   buildMultiTrackAlbumWithDifferentPerformers(),
			wantPass: true,
		},
		{
			name:     "pass - two composers only",
			actual:   buildAlbumWithNComposers(2),
			wantPass: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.AlbumArtistTag(tt.actual, tt.actual)
			
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

// buildMultiTrackAlbumWithSamePerformer creates album with same performer on all tracks
func buildMultiTrackAlbumWithSamePerformer(performerName string, trackCount int) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist(performerName, domain.RoleEnsemble)
	
	album, _ := domain.NewAlbum("Album", 1963)
	for i := range trackCount {
		track, _ := domain.NewTrack(1, i+1, fmt.Sprintf("Track %d", i+1),
			[]domain.Artist{composer, ensemble})
		album.AddTrack(track)
	}
	
	return album
}

// buildVariousComposersAlbum creates album with multiple different composers
func buildVariousComposersAlbum(composerCount int) *domain.Album {
	composers := []string{"Beethoven", "Mozart", "Bach", "Haydn", "Brahms", "Schubert"}
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	
	album, _ := domain.NewAlbum("Various Artists Album", 1963)
	for i := range composerCount {
		composer, _ := domain.NewArtist(composers[i%len(composers)], domain.RoleComposer)
		track, _ := domain.NewTrack(1, i+1, fmt.Sprintf("Work %d", i+1),
			[]domain.Artist{composer, ensemble})
		album.AddTrack(track)
	}
	
	return album
}

// buildMultiTrackAlbumWithDifferentPerformers creates album with different performers
func buildMultiTrackAlbumWithDifferentPerformers() *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	performers := []string{"Pollini", "Arrau", "Brendel", "Ashkenazy"}
	
	album, _ := domain.NewAlbum("Album", 1963)
	for i, performer := range performers {
		soloist, _ := domain.NewArtist(performer, domain.RoleSoloist)
		track, _ := domain.NewTrack(1, i+1, fmt.Sprintf("Work %d", i+1),
			[]domain.Artist{composer, soloist})
		album.AddTrack(track)
	}
	
	return album
}

// buildAlbumWithNComposers creates album with N different composers
func buildAlbumWithNComposers(n int) *domain.Album {
	composers := []string{"Beethoven", "Mozart"}
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	
	album, _ := domain.NewAlbum("Album", 1963)
	for i := 0; i < n; i++ {
		composer, _ := domain.NewArtist(composers[i%len(composers)], domain.RoleComposer)
		track, _ := domain.NewTrack(1, i+1, fmt.Sprintf("Work %d", i+1),
			[]domain.Artist{composer, ensemble})
		album.AddTrack(track)
	}
	
	return album
}
