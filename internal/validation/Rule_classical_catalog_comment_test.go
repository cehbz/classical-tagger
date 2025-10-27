package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CatalogInfoInComments(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Album
		WantPass bool
		WantInfo int
	}{
		{
			Name:     "pass - complete edition info",
			Actual:   buildAlbumWithCompleteEdition(),
			WantPass: true,
		},
		{
			Name:     "info - no edition info",
			Actual:   buildAlbumWithTitle("Symphony", "1963"),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - missing label",
			Actual:   buildAlbumWithPartialEdition("", "CAT123", 1990),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - missing catalog",
			Actual:   buildAlbumWithPartialEdition("Deutsche Grammophon", "", 1990),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - missing year",
			Actual:   buildAlbumWithPartialEdition("Deutsche Grammophon", "CAT123", 0),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "info - multiple missing fields",
			Actual:   buildAlbumWithPartialEdition("", "", 0),
			WantPass: false,
			WantInfo: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.CatalogInfoInComments(tt.Actual, tt.Actual)

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

// buildAlbumWithCompleteEdition creates album with complete edition information
func buildAlbumWithCompleteEdition() *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{composer, ensemble}}

	edition := &domain.Edition{Label: "Deutsche Grammophon", Year: 1990, CatalogNumber: "DG-479-0334"}
	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: []*domain.Track{&track}, Edition: edition}
}

// buildAlbumWithPartialEdition creates album with partial edition information
func buildAlbumWithPartialEdition(label, catalog string, year int) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{composer, ensemble}}

	edition := &domain.Edition{Label: label, Year: year, CatalogNumber: catalog}
	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: []*domain.Track{&track}, Edition: edition}
}
