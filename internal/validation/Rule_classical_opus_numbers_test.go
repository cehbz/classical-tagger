package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_OpusNumbers(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name      string
		actual    *domain.Album
		reference *domain.Album
		wantPass  bool
		wantInfo  int
	}{
		{
			name:     "valid - has opus number",
			actual:   buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			wantPass: true,
		},
		{
			name:     "info - missing opus number for Beethoven",
			actual:   buildAlbumWithTrackTitleAndComposer("Symphony No. 5", "Beethoven"),
			wantPass: false,
			wantInfo: 1,
		},
		{
			name:     "valid - has BWV number",
			actual:   buildAlbumWithTrackTitleAndComposer("Fugue in D Minor, BWV 1080", "Bach"),
			wantPass: true,
		},
		{
			name:     "valid - has Köchel number",
			actual:   buildAlbumWithTrackTitleAndComposer("Symphony No. 40, K. 550", "Mozart"),
			wantPass: true,
		},
		{
			name:     "valid - has Hoboken number",
			actual:   buildAlbumWithTrackTitleAndComposer("Sonata, Hob. XVI:52", "Haydn"),
			wantPass: true,
		},
		{
			name:     "valid - has Deutsch number",
			actual:   buildAlbumWithTrackTitleAndComposer("Impromptu, D. 899", "Schubert"),
			wantPass: true,
		},
		{
			name:     "valid - has RV number",
			actual:   buildAlbumWithTrackTitleAndComposer("Concerto, RV 315", "Vivaldi"),
			wantPass: true,
		},
		{
			name:     "pass - no catalog system for this composer",
			actual:   buildAlbumWithTrackTitleAndComposer("Symphony", "Contemporary Composer"),
			wantPass: true,
		},
		{
			name:      "info - reference has opus but actual doesn't",
			actual:    buildAlbumWithTrackTitleAndComposer("Symphony No. 5", "Beethoven"),
			reference: buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			wantPass:  false,
			wantInfo:  1,
		},
		{
			name:      "pass - both have opus",
			actual:    buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			reference: buildAlbumWithTrackTitleAndComposer("Symphony No. 5, Op. 67", "Beethoven"),
			wantPass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.OpusNumbers(tt.actual, tt.reference)

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

func TestHasOpusNumber(t *testing.T) {
	tests := []struct {
		title string
		want  bool
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
		t.Run(tt.title, func(t *testing.T) {
			got := hasOpusNumber(tt.title)
			if got != tt.want {
				t.Errorf("hasOpusNumber(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}

func TestNeedsCatalogNumber(t *testing.T) {
	tests := []struct {
		composer string
		want     bool
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
		t.Run(tt.composer, func(t *testing.T) {
			got := needsCatalogNumber(tt.composer)
			if got != tt.want {
				t.Errorf("needsCatalogNumber(%q) = %v, want %v", tt.composer, got, tt.want)
			}
		})
	}
}

// buildAlbumWithTrackTitleAndComposer creates album with specific track title and composer
func buildAlbumWithTrackTitleAndComposer(trackTitle, composerName string) *domain.Album {
	composer, _ := domain.NewArtist(composerName, domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, trackTitle, []domain.Artist{composer, ensemble})
	album, _ := domain.NewAlbum("Album", 1963)
	album.AddTrack(track)
	return album
}
