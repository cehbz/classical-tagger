package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArtistPositionInFilename(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Torrent
		WantPass   bool
		WantIssues int
	}{
		{
			Name: "valid - single composer album (position irrelevant)",
			Actual: buildSingleComposerTorrentWithFilenames(
				"Beethoven",
				"01 - Symphony No. 1.flac",
				"02 - Symphony No. 2.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid - multi-composer, artist after track number",
			Actual: buildMultiComposerTorrentWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Bach - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02 - Vivaldi - Concerto.flac"}},
			),
			WantPass: true,
		},
		{
			Name: "invalid - multi-composer, artist before track number",
			Actual: buildMultiComposerTorrentWithFilenames(
				[]composerTrack{{"Bach", 1, "Bach - 01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "Vivaldi - 02 - Concerto.flac"}},
			),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name: "valid - multi-composer, no artist in filename",
			Actual: buildMultiComposerTorrentWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02 - Concerto.flac"}},
			),
			WantPass: true,
		},
		{
			Name: "invalid - some tracks have artist before number",
			Actual: buildMultiComposerTorrentWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "Vivaldi - 02 - Concerto.flac"}},
			),
			WantPass:   false,
			WantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			issues := []domain.ValidationIssue{}
			for _, track := range tt.Actual.Tracks() {
				result := rules.ArtistPositionInFilename(track, nil, tt.Actual, nil)
				issues = append(issues, result.Issues...)
			}

			if len(issues) != tt.WantIssues {
				t.Errorf("Issues = %d, want %d", len(issues), tt.WantIssues)
				for _, issue := range issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}

func TestIsMultiComposerAlbum(t *testing.T) {
	tests := []struct {
		Name      string
		Album     *domain.Torrent
		WantMulti bool
	}{
		{
			Name:      "single composer",
			Album:     buildSingleComposerTorrentWithFilenames("Bach", "01.flac", "02.flac"),
			WantMulti: false,
		},
		{
			Name: "multiple composers",
			Album: buildMultiComposerTorrentWithFilenames(
				[]composerTrack{{"Bach", 1, "01.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02.flac"}},
			),
			WantMulti: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := isMultiComposerAlbum(tt.Album)
			if got != tt.WantMulti {
				t.Errorf("isMultiComposerAlbum() = %v, want %v", got, tt.WantMulti)
			}
		})
	}
}

func TestContainsArtistName(t *testing.T) {
	bach := domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer}
	vivaldi := domain.Artist{Name: "Antonio Vivaldi", Role: domain.RoleComposer}

	tests := []struct {
		Name     string
		Filename string
		Artists  []domain.Artist
		Want     bool
	}{
		{
			Name:     "last name before track number",
			Filename: "Bach - 01 - Prelude.flac",
			Artists:  []domain.Artist{bach},
			Want:     true,
		},
		{
			Name:     "last name after track number",
			Filename: "01 - Bach - Prelude.flac",
			Artists:  []domain.Artist{bach},
			Want:     false,
		},
		{
			Name:     "full name before track number",
			Filename: "Antonio Vivaldi - 01 - Concerto.flac",
			Artists:  []domain.Artist{vivaldi},
			Want:     true,
		},
		{
			Name:     "no artist in filename",
			Filename: "01 - Prelude.flac",
			Artists:  []domain.Artist{bach},
			Want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := containsArtistName(tt.Filename, tt.Artists)
			if got != tt.Want {
				t.Errorf("containsArtistName(%q) = %v, want %v", tt.Filename, got, tt.Want)
			}
		})
	}
}
