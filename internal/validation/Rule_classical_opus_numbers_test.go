package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_OpusNumbers(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name      string
		Actual    *domain.Album
		Reference *domain.Album
		WantPass  bool
		WantInfo  int
	}{
		{
			Name:     "valid - has opus number",
			Actual:   buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			WantPass: true,
		},
		{
			Name:     "info - missing opus number for Beethoven",
			Actual:   buildAlbumWithTrackTitleAndComposer("Symphony No. 5", "Beethoven"),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - has BWV number",
			Actual:   buildAlbumWithTrackTitleAndComposer("Fugue in D Minor, BWV 1080", "Bach"),
			WantPass: true,
		},
		{
			Name:     "valid - has Köchel number",
			Actual:   buildAlbumWithTrackTitleAndComposer("Symphony No. 40, K. 550", "Mozart"),
			WantPass: true,
		},
		{
			Name:     "valid - has Hoboken number",
			Actual:   buildAlbumWithTrackTitleAndComposer("Sonata, Hob. XVI:52", "Haydn"),
			WantPass: true,
		},
		{
			Name:     "valid - has Deutsch number",
			Actual:   buildAlbumWithTrackTitleAndComposer("Impromptu, D. 899", "Schubert"),
			WantPass: true,
		},
		{
			Name:     "valid - has RV number",
			Actual:   buildAlbumWithTrackTitleAndComposer("Concerto, RV 315", "Vivaldi"),
			WantPass: true,
		},
		{
			Name:     "pass - no catalog system for this composer",
			Actual:   buildAlbumWithTrackTitleAndComposer("Symphony", "Contemporary Composer"),
			WantPass: true,
		},
		{
			Name:      "info - reference has opus but actual doesn't",
			Actual:    buildAlbumWithTrackTitleAndComposer("Symphony No. 5", "Beethoven"),
			Reference: buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "pass - both have opus",
			Actual:    buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			Reference: buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			WantPass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.OpusNumbers(tt.Actual, tt.Reference)

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

func TestHasOpusNumber(t *testing.T) {
	tests := []struct {
		Title string
		Want  bool
	}{
		{"Symphony No. 5, Op. 67", true},
		{"Symphony No. 5, Op 67", true},
		{"Fugue in D Minor, BWV 1080", true},
		{"Symphony No. 40, K. 550", true},
		{"Symphony No. 40, K 550", true},
		{"Sonata, Hob. XVI:52", true},
		{"Impromptu, D. 899", true},
		{"Concerto, RV 315", true},
		{"Symphony No. 5", false},
		{"Piano Concerto", false},
	}

	for _, tt := range tests {
		t.Run(tt.Title, func(t *testing.T) {
			got := hasOpusNumber(tt.Title)
			if got != tt.Want {
				t.Errorf("hasOpusNumber(%q) = %v, want %v", tt.Title, got, tt.Want)
			}
		})
	}
}

func TestNeedsCatalogNumber(t *testing.T) {
	tests := []struct {
		Composer string
		Want     bool
	}{
		{"Beethoven", true},
		{"Ludwig van Beethoven", true},
		{"Mozart", true},
		{"Wolfgang Amadeus Mozart", true},
		{"Bach", true},
		{"Johann Sebastian Bach", true},
		{"Haydn", true},
		{"Schubert", true},
		{"Vivaldi", true},
		{"Brahms", true},
		{"Contemporary Composer", false},
		{"John Williams", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.Composer, func(t *testing.T) {
			got := needsCatalogNumber(tt.Composer)
			if got != tt.Want {
				t.Errorf("needsCatalogNumber(%q) = %v, want %v", tt.Composer, got, tt.Want)
			}
		})
	}
}

// buildAlbumWithTrackTitleAndComposer creates album with specific track title and composer
func buildAlbumWithTrackTitleAndComposer(trackTitle, composerName string) *domain.Album {
	composer := domain.Artist{Name: composerName, Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	track := &domain.Track{Disc: 1, Track: 1, Title: trackTitle, Artists: []domain.Artist{composer, ensemble}}
	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: []*domain.Track{track}}
}
