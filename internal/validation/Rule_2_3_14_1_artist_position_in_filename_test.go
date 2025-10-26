package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArtistPositionInFilename(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - single composer album (position irrelevant)",
			actual: buildSingleComposerAlbumWithFilenames(
				"Beethoven",
				"01 - Symphony No. 1.flac",
				"02 - Symphony No. 2.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - multi-composer, artist after track number",
			actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Bach - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02 - Vivaldi - Concerto.flac"}},
			),
			wantPass: true,
		},
		{
			name: "invalid - multi-composer, artist before track number",
			actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "Bach - 01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "Vivaldi - 02 - Concerto.flac"}},
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "valid - multi-composer, no artist in filename",
			actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02 - Concerto.flac"}},
			),
			wantPass: true,
		},
		{
			name: "invalid - some tracks have artist before number",
			actual: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "01 - Prelude.flac"}},
				[]composerTrack{{"Vivaldi", 2, "Vivaldi - 02 - Concerto.flac"}},
			),
			wantPass:   false,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ArtistPositionInFilename(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass && len(result.Issues()) != tt.wantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues()), tt.wantIssues)
				for _, issue := range result.Issues() {
					t.Logf("  Issue: %s", issue.Message())
				}
			}
		})
	}
}

func TestIsMultiComposerAlbum(t *testing.T) {
	tests := []struct {
		name      string
		album     *domain.Album
		wantMulti bool
	}{
		{
			name:      "single composer",
			album:     buildSingleComposerAlbumWithFilenames("Bach", "01.flac", "02.flac"),
			wantMulti: false,
		},
		{
			name: "multiple composers",
			album: buildMultiComposerAlbumWithFilenames(
				[]composerTrack{{"Bach", 1, "01.flac"}},
				[]composerTrack{{"Vivaldi", 2, "02.flac"}},
			),
			wantMulti: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMultiComposerAlbum(tt.album)
			if got != tt.wantMulti {
				t.Errorf("isMultiComposerAlbum() = %v, want %v", got, tt.wantMulti)
			}
		})
	}
}

func TestContainsArtistName(t *testing.T) {
	bach, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	vivaldi, _ := domain.NewArtist("Antonio Vivaldi", domain.RoleComposer)

	tests := []struct {
		name     string
		filename string
		artists  []domain.Artist
		want     bool
	}{
		{
			name:     "last name before track number",
			filename: "Bach - 01 - Prelude.flac",
			artists:  []domain.Artist{bach},
			want:     true,
		},
		{
			name:     "last name after track number",
			filename: "01 - Bach - Prelude.flac",
			artists:  []domain.Artist{bach},
			want:     false,
		},
		{
			name:     "full name before track number",
			filename: "Antonio Vivaldi - 01 - Concerto.flac",
			artists:  []domain.Artist{vivaldi},
			want:     true,
		},
		{
			name:     "no artist in filename",
			filename: "01 - Prelude.flac",
			artists:  []domain.Artist{bach},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsArtistName(tt.filename, tt.artists)
			if got != tt.want {
				t.Errorf("containsArtistName(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// composerTrack represents a track with specific composer
type composerTrack struct {
	composerName string
	trackNum     int
	filename     string
}

// buildSingleComposerAlbumWithFilenames creates album with one composer
func buildSingleComposerAlbumWithFilenames(composerName string, filenames ...string) *domain.Album {
	composer, _ := domain.NewArtist(composerName, domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	artists := []domain.Artist{composer, ensemble}

	album, _ := domain.NewAlbum("Album", 1963)
	for i, filename := range filenames {
		track, _ := domain.NewTrack(1, i+1, "Work "+string(rune('A'+i)), artists)
		track = track.WithName(filename)
		album.AddTrack(track)
	}
	return album
}

// buildMultiComposerAlbumWithFilenames creates album with multiple composers
func buildMultiComposerAlbumWithFilenames(composerTracks ...[]composerTrack) *domain.Album {
	album, _ := domain.NewAlbum("Various Composers", 1963)
	for _, ctList := range composerTracks {
		for _, ct := range ctList {
			composer, _ := domain.NewArtist(ct.composerName, domain.RoleComposer)
			ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
			artists := []domain.Artist{composer, ensemble}

			track, _ := domain.NewTrack(1, ct.trackNum, "Work", artists)
			track = track.WithName(ct.filename)
			album.AddTrack(track)
		}
	}

	return album
}
