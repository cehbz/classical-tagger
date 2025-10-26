package validation

import (
	"fmt"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CapitalizationTrump(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name      string
		actual    *domain.Album
		reference *domain.Album
		wantPass  bool
		wantInfo  int
	}{
		{
			name:     "pass - no reference",
			actual:   buildAlbumWithTitle("Symphony No. 5", "1963"),
			wantPass: true,
		},
		{
			name:      "info - no improvement",
			actual:    buildAlbumWithBadCaps(),
			reference: buildAlbumWithGoodCaps(),
			wantPass:  false,
			wantInfo:  1,
		},
		{
			name:      "info - significant improvement",
			actual:    buildAlbumWithGoodCaps(),
			reference: buildAlbumWithBadCaps(),
			wantPass:  false,
			wantInfo:  1,
		},
		{
			name:      "info - same quality",
			actual:    buildAlbumWithGoodCaps(),
			reference: buildAlbumWithGoodCaps(),
			wantPass:  false,
			wantInfo:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CapitalizationTrump(tt.actual, tt.reference)

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

func TestCountCapitalizationIssues(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name      string
		album     *domain.Album
		wantCount int
	}{
		{
			name:      "good capitalization",
			album:     buildAlbumWithGoodCaps(),
			wantCount: 0,
		},
		{
			name:      "bad capitalization",
			album:     buildAlbumWithBadCaps(),
			wantCount: 4, // Title + 3 track titles
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := rules.countCapitalizationIssues(tt.album)

			if count != tt.wantCount {
				t.Errorf("countCapitalizationIssues() = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

// buildAlbumWithGoodCaps creates album with proper capitalization
func buildAlbumWithGoodCaps() *domain.Album {
	composer, _ := domain.NewArtist("Ludwig van Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Berlin Philharmonic", domain.RoleEnsemble)

	tracks := make([]*domain.Track, 3)
	for i := 0; i < 3; i++ {
		track, _ := domain.NewTrack(1, i+1,
			fmt.Sprintf("Symphony No. %d", i+1),
			[]domain.Artist{composer, ensemble})
		tracks[i] = track
	}

	album, _ := domain.NewAlbum("Beethoven - Symphonies [1963] [FLAC]", 1963)
	for _, track := range tracks {
		album.AddTrack(track)
	}
	return album
}

// buildAlbumWithBadCaps creates album with poor capitalization
func buildAlbumWithBadCaps() *domain.Album {
	composer, _ := domain.NewArtist("BEETHOVEN", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("berlin philharmonic", domain.RoleEnsemble)

	tracks := make([]*domain.Track, 3)
	for i := 0; i < 3; i++ {
		track, _ := domain.NewTrack(1, i+1,
			fmt.Sprintf("SYMPHONY NO. %d", i+1),
			[]domain.Artist{composer, ensemble})
		tracks[i] = track
	}

	album, _ := domain.NewAlbum("BEETHOVEN - SYMPHONIES", 1963)
	for _, track := range tracks {
		album.AddTrack(track)
	}
	return album
}
