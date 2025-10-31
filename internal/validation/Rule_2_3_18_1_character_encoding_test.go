package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_AlbumCharacterEncoding(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantErrors int
	}{
		{
			Name:     "valid - proper UTF-8",
			Actual:   NewAlbum().WithTitle("Beethoven - Symphony No. 5").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper accented characters",
			Actual:   NewAlbum().WithTitle("Dvořák - String Quartet").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper German umlauts",
			Actual:   NewAlbum().WithTitle("Arnold Schönberg - Symphonies").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:       "error - mojibake pattern Ã©",
			Actual:     NewAlbum().WithTitle("ConcertÃ© in D").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - mojibake pattern â€™",
			Actual:     NewAlbum().WithTitle("Donâ€™t Stop Believin").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - replacement character",
			Actual:     NewAlbum().WithTitle("Concert\uFFFD in D").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - null byte",
			Actual:     NewAlbum().WithTitle("Concert\x00 in D").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - proper Russian characters",
			Actual:   NewAlbum().WithTitle("Tchaikovsky - Лебединое озеро").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:       "error - encoding issue in artist name",
			Actual:     NewAlbum().WithTitle("DvÃ¶rÃ¡k").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Logf("debug album title: %q", tt.Actual.Title)
			result := rules.AlbumCharacterEncoding(tt.Actual, nil)
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

func TestRules_TrackCharacterEncoding(t *testing.T) {
	rules := NewRules()

	// builder sanity check
	t.Run("builder_sanity_for_track_title", func(t *testing.T) {
		album := NewAlbum().ClearTracks().AddTrack().WithTitle("ZZZ").Build().Build()
		if len(album.Tracks) != 1 || album.Tracks[0].Title != "ZZZ" {
			t.Fatalf("builder produced title=%q (tracks=%d), want ZZZ,1", album.Tracks[0].Title, len(album.Tracks))
		}
	})

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantErrors int
	}{
		{
			Name:     "valid - proper UTF-8",
			Actual:   NewAlbum().WithTitle("Beethoven - Symphony No. 5").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper accented characters",
			Actual:   NewAlbum().WithTitle("Dvořák - String Quartet").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper German umlauts",
			Actual:   NewAlbum().WithTitle("Arnold Schönberg - Symphonies").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:       "error - mojibake pattern Ã©",
			Actual:     buildAlbumWithTrackTitle("ConcertÃ© in D"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - mojibake pattern â€™",
			Actual:     buildAlbumWithTrackTitle("Donâ€™t Stop Believin"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - replacement character",
			Actual:     buildAlbumWithTrackTitle("Concert\uFFFD in D"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - null byte",
			Actual:     buildAlbumWithTrackTitle("Concert\x00 in D"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - proper Russian characters",
			Actual:   NewAlbum().WithTitle("Tchaikovsky - Лебединое озеро").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
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
		for _, track := range tt.Actual.Tracks {
			t.Run(tt.Name, func(t *testing.T) {
				t.Logf("debug tracks=%d firstTitle=%q trackTitle=%q filename=%q", len(tt.Actual.Tracks), tt.Actual.Tracks[0].Title, track.Title, track.Name)
				result := rules.TrackCharacterEncoding(track, nil, nil, nil)

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
