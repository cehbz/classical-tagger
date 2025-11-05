package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_AlbumCharacterEncoding(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Torrent
		WantPass   bool
		WantErrors int
	}{
		{
			Name:     "valid - proper UTF-8",
			Actual:   NewTorrent().WithTitle("Beethoven - Symphony No. 5").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper accented characters",
			Actual:   NewTorrent().WithTitle("Dvořák - String Quartet").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper German umlauts",
			Actual:   NewTorrent().WithTitle("Arnold Schönberg - Symphonies").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:       "error - mojibake pattern Ã©",
			Actual:     NewTorrent().WithTitle("ConcertÃ© in D").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - mojibake pattern â€™",
			Actual:     NewTorrent().WithTitle("Donâ€™t Stop Believin").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - replacement character",
			Actual:     NewTorrent().WithTitle("Concert\uFFFD in D").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - null byte",
			Actual:     NewTorrent().WithTitle("Concert\x00 in D").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - proper Russian characters",
			Actual:   NewTorrent().WithTitle("Tchaikovsky - Лебединое озеро").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:       "error - encoding issue in artist name",
			Actual:     NewTorrent().WithTitle("DvÃ¶rÃ¡k").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
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
		album := NewTorrent().ClearTracks().AddTrack().WithTitle("ZZZ").Build().Build()
		tracks := album.Tracks()
		if len(tracks) != 1 {
			t.Fatalf("builder produced %d tracks, want 1", len(tracks))
		}
		if tracks[0].Title != "ZZZ" {
			t.Fatalf("builder produced title=%q, want ZZZ", tracks[0].Title)
		}
	})

	tests := []struct {
		Name       string
		Actual     *domain.Torrent
		WantPass   bool
		WantErrors int
	}{
		{
			Name:     "valid - proper UTF-8",
			Actual:   NewTorrent().WithTitle("Beethoven - Symphony No. 5").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper accented characters",
			Actual:   NewTorrent().WithTitle("Dvořák - String Quartet").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - proper German umlauts",
			Actual:   NewTorrent().WithTitle("Arnold Schönberg - Symphonies").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:       "error - mojibake pattern Ã©",
			Actual:     buildTorrentWithTrackTitle("ConcertÃ© in D"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - mojibake pattern â€™",
			Actual:     buildTorrentWithTrackTitle("Donâ€™t Stop Believin"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - replacement character",
			Actual:     buildTorrentWithTrackTitle("Concert\uFFFD in D"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - null byte",
			Actual:     buildTorrentWithTrackTitle("Concert\x00 in D"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:     "valid - proper Russian characters",
			Actual:   NewTorrent().WithTitle("Tchaikovsky - Лебединое озеро").WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).Build(),
			WantPass: true,
		},
		{
			Name:       "error - encoding issue in artist name",
			Actual:     buildTorrentWithArtistName("DvÃ¶rÃ¡k"),
			WantPass:   false,
			WantErrors: 1,
		},
	}

	for _, tt := range tests {
		tracks := tt.Actual.Tracks()
		for _, track := range tracks {
			t.Run(tt.Name, func(t *testing.T) {
				t.Logf("debug tracks=%d firstTitle=%q trackTitle=%q filename=%q", len(tracks), tracks[0].Title, track.Title, track.File.Path)
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
