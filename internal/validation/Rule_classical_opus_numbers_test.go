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
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5, Op. 67").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "info - missing opus number for Beethoven",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - has BWV number",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Fugue in D Minor, BWV 1080").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - has Köchel number",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 40, K. 550").ClearArtists().WithArtists(domain.Artist{Name: "Mozart", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - has Hoboken number",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Sonata, Hob. XVI:52").ClearArtists().WithArtists(domain.Artist{Name: "Haydn", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - has Deutsch number",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Impromptu, D. 899").ClearArtists().WithArtists(domain.Artist{Name: "Schubert", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - has RV number",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Concerto, RV 315").ClearArtists().WithArtists(domain.Artist{Name: "Vivaldi", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "pass - no catalog system for this composer",
			Actual:   NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Contemporary Composer", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:      "info - reference has opus but actual doesn't",
			Actual:    NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			Reference: NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5, Op. 67").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass:  false,
			WantInfo:  1,
		},
		{
			Name:      "pass - both have opus",
			Actual:    NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5, Op. 67").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			Reference: NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5, Op. 67").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass:  true,
		},
	}

	for _, tt := range tests {
		for i, track := range tt.Actual.Tracks {
			var refTrack *domain.Track
			if tt.Reference != nil {
				refTrack = tt.Reference.Tracks[i]
			}
			t.Run(tt.Name, func(t *testing.T) {
				result := rules.OpusNumbers(track, refTrack, nil, nil)

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
}

func TestExtractOpusNumber(t *testing.T) {
	tests := []struct {
		Title string
		Want  string
	}{
		{"Symphony No. 5, Op. 67", "Op. 67"},
		{"Symphony No. 5, Op 67", "Op 67"},
		{"Fugue in D Minor, BWV 1080", "BWV 1080"},
		{"Symphony No. 40, K. 550", "K. 550"},
		{"Symphony No. 40, K 550", "K 550"},
		{"Sonata, Hob. XVI:52", "Hob. XVI:52"},
		{"Impromptu, D. 899", "D. 899"},
		{"Concerto, RV 315", "RV 315"},
		{"Symphony No. 5", ""},
		{"Piano Concerto", ""},
	}

	for _, tt := range tests {
		t.Run(tt.Title, func(t *testing.T) {
			got := extractOpusNumber(tt.Title)
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
