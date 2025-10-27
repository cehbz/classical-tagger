package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArtistPositionInFilename(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name: "valid - single composer album (position irrelevant)",
			Actual: buildSingleComposerAlbumWithFilenames(
				"Beethoven",
				"01 - Symphony No. 1.flac",
				"02 - Symphony No. 2.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid - multi-composer, artist after track number",
			Actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Bach - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02 - Vivaldi - Concerto.flac"}},
			),
			WantPass: true,
		},
		{
			Name: "invalid - multi-composer, artist before track number",
			Actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "Bach - 01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "Vivaldi - 02 - Concerto.flac"}},
			),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name: "valid - multi-composer, no artist in filename",
			Actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02 - Concerto.flac"}},
			),
			WantPass: true,
		},
		{
			Name: "invalid - some tracks have artist before number",
			Actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "Vivaldi - 02 - Concerto.flac"}},
			),
			WantPass:   false,
			WantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.ArtistPositionInFilename(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass && len(result.Issues) != tt.WantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues), tt.WantIssues)
				for _, issue := range result.Issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}

func TestIsMultiComposerAlbum(t *testing.T) {
	tests := []struct {
		Name      string
		Album     *domain.Album
		WantMulti bool
	}{
		{
			Name:      "single composer",
			Album:     buildSingleComposerAlbumWithFilenames("Bach", "01.flac", "02.flac"),
			WantMulti: false,
		},
		{
			Name: "multiple composers",
			Album: buildMultiComposerAlbumWithFilenames(
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

// composerTrack represents a track with specific composer
type composerTrack struct {
	ComposerName string
	TrackNum     int
	Filename     string
}

// buildSingleComposerAlbumWithFilenames creates album with one composer
func buildSingleComposerAlbumWithFilenames(composerName string, filenames ...string) *domain.Album {
	tracks := make([]*domain.Track, len(filenames))

	for i, _ := range filenames {
		tracks[i] = &domain.Track{
			Disc:  1,
			Track: i + 1,
			Title: "Work " + string(rune('A'+i)),
			Artists: []domain.Artist{
				domain.Artist{Name: composerName, Role: domain.RoleComposer},
				domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
			},
		}
	}
	return &domain.Album{
		Title:        "Album",
		OriginalYear: 1963,
		Tracks:       tracks,
	}
}

// buildMultiComposerAlbumWithFilenames creates album with multiple composers
func buildMultiComposerAlbumWithFilenames(composerTracks ...[]composerTrack) *domain.Album {
	tracks := make([]*domain.Track, 0)
	for _, ctList := range composerTracks {
		for _, ct := range ctList {
			tracks = append(tracks, &domain.Track{
				Disc:  1,
				Track: ct.TrackNum,
				Title: "Work " + string(rune('A'+ct.TrackNum)),
				Name:  ct.Filename,
				Artists: []domain.Artist{
					domain.Artist{Name: ct.ComposerName, Role: domain.RoleComposer},
					domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
				},
			})
		}
	}
	return &domain.Album{
		Title:        "Various Composers",
		OriginalYear: 1963,
		Tracks:       tracks,
	}
}
