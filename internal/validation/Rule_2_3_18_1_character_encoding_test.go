package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_CharacterEncoding(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantErrors int
	}{
		{
			name:     "valid - proper UTF-8",
			actual:   buildAlbumWithTitle("Beethoven - Symphony No. 5", "1963"),
			wantPass: true,
		},
		{
			name:     "valid - proper accented characters",
			actual:   buildAlbumWithTitle("Dvořák - String Quartet", "1963"),
			wantPass: true,
		},
		{
			name:     "valid - proper German umlauts",
			actual:   buildAlbumWithTitle("Bruckner - Symphonies", "1963"),
			wantPass: true,
		},
		{
			name:       "error - mojibake pattern Ã©",
			actual:     buildAlbumWithTitle("ConcertÃ© in D", "1963"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - mojibake pattern â€™",
			actual:     buildAlbumWithTitle("Donâ€™t Stop Believin", "1963"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - replacement character",
			actual:     buildAlbumWithTitle("Concert\uFFFD in D", "1963"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - null byte",
			actual:     buildAlbumWithTitle("Concert\x00 in D", "1963"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:     "valid - proper Russian characters",
			actual:   buildAlbumWithTitle("Tchaikovsky - Лебединое озеро", "1963"),
			wantPass: true,
		},
		{
			name:       "error - encoding issue in artist name",
			actual:     buildAlbumWithArtistName("DvÃ¶rÃ¡k"),
			wantPass:   false,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CharacterEncoding(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				errorCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelError {
						errorCount++
					}
				}

				if errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

func TestHasEncodingIssues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
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
		t.Run(tt.name, func(t *testing.T) {
			got := hasEncodingIssues(tt.input)
			if got != tt.want {
				t.Errorf("hasEncodingIssues(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// buildAlbumWithArtistName creates album with specific artist name
func buildAlbumWithArtistName(artistName string) *domain.Album {
	composer, _ := domain.NewArtist(artistName, domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})
	album, _ := domain.NewAlbum("Album", 1963)
	album.AddTrack(track)
	return album
}
