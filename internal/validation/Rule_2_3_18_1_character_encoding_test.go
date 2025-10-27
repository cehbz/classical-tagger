package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CharacterEncoding(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantErrors int
	}{
		{
			Name:     "valid - proper UTF-8",
			Actual:   buildAlbumWithTitle("Beethoven - Symphony No. 5", "1963"),
			WantPass: true,
		},
		{
			Name:     "valid - proper accented characters",
			Actual:   buildAlbumWithTitle("Dvořák - String Quartet", "1963"),
			WantPass: true,
		},
		{
			Name:     "valid - proper German umlauts",
			Actual:   buildAlbumWithTitle("Bruckner - Symphonies", "1963"),
			WantPass: true,
		},
		{
			Name:       "error - mojibake pattern Ã©",
			Actual:     buildAlbumWithTitle("ConcertÃ© in D", "1963"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - mojibake pattern â€™",
			Actual:     buildAlbumWithTitle("Donâ€™t Stop Believin", "1963"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - replacement character",
			Actual:     buildAlbumWithTitle("Concert\uFFFD in D", "1963"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - null byte",
			Actual:     buildAlbumWithTitle("Concert\x00 in D", "1963"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - proper Russian characters",
			Actual:   buildAlbumWithTitle("Tchaikovsky - Лебединое озеро", "1963"),
			WantPass: true,
		},
		{
			Name:       "error - encoding issue in artist name",
			Actual:     buildAlbumWithArtistName("DvÃ¶rÃ¡k"),
			WantPass:   false,
			WantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.CharacterEncoding(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelError {
						errorCount++
					}
				}

				if errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

func TestHasEncodingIssues(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  bool
	}{
		{"valid ASCII", "Symphony No. 5", false},
		{"valid UTF-8 accents", "Dvořák", false},
		{"valid UTF-8 umlauts", "Brückner", false},
		{"valid Russian", "Tchaikovsky", false},
		{"mojibake Ã©", "CafÃ©", true},
		{"mojibake Ã¨", "crÃ¨me", true},
		{"mojibake smart quote", "donâ€™t", true},
		{"replacement char", "test\uFFFDing", true},
		{"null byte", "test\x00ing", true},
		{"control char", "test\x01ing", true},
		{"valid newline", "line1\nline2", false},
		{"valid tab", "col1\tcol2", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := hasEncodingIssues(tt.Input)
			if got != tt.Want {
				t.Errorf("hasEncodingIssues(%q) = %v, want %v", tt.Input, got, tt.Want)
			}
		})
	}
}

// buildAlbumWithArtistName creates album with specific artist name
func buildAlbumWithArtistName(artistName string) *domain.Album {
	composer := domain.Artist{Name: artistName, Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{composer, ensemble}}
	return &domain.Album{Title: "Album", OriginalYear: 1963, Tracks: []*domain.Track{&track}}
}
