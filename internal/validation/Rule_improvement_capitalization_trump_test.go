package validation

import (
	"fmt"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CapitalizationTrump(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name      string
		Actual    *domain.Album
		Reference *domain.Album
		WantPass  bool
		WantInfo  int
	}{
		{
			Name:     "pass - no reference",
			Actual:   buildAlbumWithTitle("Symphony No. 5", "1963"),
			WantPass: true,
		},
		{
			Name:      "info - no improvement",
			Actual:    buildAlbumWithBadCaps(),
			Reference: buildAlbumWithGoodCaps(),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "info - significant improvement",
			Actual:    buildAlbumWithGoodCaps(),
			Reference: buildAlbumWithBadCaps(),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "info - same quality",
			Actual:    buildAlbumWithGoodCaps(),
			Reference: buildAlbumWithGoodCaps(),
			WantPass:  false,
			WantInfo:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.CapitalizationTrump(tt.Actual, tt.Reference)

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

func TestCountCapitalizationIssues(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name      string
		Album     *domain.Album
		WantCount int
	}{
		{
			Name:      "good capitalization",
			Album:     buildAlbumWithGoodCaps(),
			WantCount: 0,
		},
		{
			Name:      "bad capitalization",
			Album:     buildAlbumWithBadCaps(),
			WantCount: 4, // Title + 3 track titles
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			count := rules.countCapitalizationIssues(tt.Album)

			if count != tt.WantCount {
				t.Errorf("countCapitalizationIssues() = %d, want %d", count, tt.WantCount)
			}
		})
	}
}

// buildAlbumWithGoodCaps creates album with proper capitalization
func buildAlbumWithGoodCaps() *domain.Album {
	composer := domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}

	tracks := make([]*domain.Track, 3)
	for i := 0; i < 3; i++ {
		tracks[i] = &domain.Track{
			Disc:    1,
			Track:   i + 1,
			Title:   fmt.Sprintf("Symphony No. %d", i+1),
			Artists: []domain.Artist{composer, ensemble},
		}
	}

	return &domain.Album{
		Title:        "Beethoven - Symphonies [1963] [FLAC]",
		OriginalYear: 1963,
		Tracks:       tracks,
	}
}

// buildAlbumWithBadCaps creates album with poor capitalization
func buildAlbumWithBadCaps() *domain.Album {
	composer := domain.Artist{Name: "BEETHOVEN", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "berlin philharmonic", Role: domain.RoleEnsemble}

	tracks := make([]*domain.Track, 3)
	for i := 0; i < 3; i++ {
		tracks[i] = &domain.Track{
			Disc:    1,
			Track:   i + 1,
			Title:   fmt.Sprintf("SYMPHONY NO. %d", i+1),
			Artists: []domain.Artist{composer, ensemble},
		}
	}

	return &domain.Album{
		Title:        "BEETHOVEN - SYMPHONIES",
		OriginalYear: 1963,
		Tracks:       tracks,
	}
}
