package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerInFolderName(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		albumTitle   string
		composerName string
		wantPass     bool
		wantWarnings int
		wantInfo     int
	}{
		{
			name:         "valid - full composer name in title",
			albumTitle:   "Ludwig van Beethoven - Symphony No. 5 [1963] [FLAC]",
			composerName: "Ludwig van Beethoven",
			wantPass:     true,
		},
		{
			name:         "valid - abbreviated composer name",
			albumTitle:   "J.S. Bach - Brandenburg Concertos [1982] [FLAC]",
			composerName: "Johann Sebastian Bach",
			wantPass:     true,
		},
		{
			name:         "info - last name only",
			albumTitle:   "Beethoven - Symphony No. 5 [1963] [FLAC]",
			composerName: "Ludwig van Beethoven",
			wantPass:     false,
			wantInfo:     1,
		},
		{
			name:         "warning - composer missing",
			albumTitle:   "Symphony No. 5 [1963] [FLAC]",
			composerName: "Ludwig van Beethoven",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "valid - Various Artists",
			albumTitle:   "Various Artists - Classical Favorites [2000] [FLAC]",
			composerName: "Bach",
			wantPass:     true, // Various artists is an exception
		},
		{
			name:         "valid - composer with particles",
			albumTitle:   "Beethoven - Piano Sonatas [1990] [FLAC]",
			composerName: "Ludwig van Beethoven",
			wantPass:     false,
			wantInfo:     1, // Has "Beethoven" but not full name
		},
		{
			name:         "valid - full name present",
			albumTitle:   "Wolfgang Amadeus Mozart - Piano Concertos [1985] [FLAC]",
			composerName: "Wolfgang Amadeus Mozart",
			wantPass:     true,
		},
		{
			name:         "valid - W.A. Mozart abbreviation",
			albumTitle:   "W.A. Mozart - Piano Concertos [1985] [FLAC]",
			composerName: "Wolfgang Amadeus Mozart",
			wantPass:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildSingleComposerAlbumWithTitle(tt.composerName, tt.albumTitle)
			result := rules.ComposerInFolderName(actual, actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelWarning {
						warningCount++
					} else if issue.Level() == domain.LevelInfo {
						infoCount++
					}
				}

				if tt.wantWarnings > 0 && warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}
				if tt.wantInfo > 0 && infoCount != tt.wantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.wantInfo)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

// buildSingleComposerAlbumWithTitle creates an album with specific title and composer
func buildSingleComposerAlbumWithTitle(composerName, albumTitle string) *domain.Album {
	composer, _ := domain.NewArtist(composerName, domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Vienna Philharmonic", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Herbert von Karajan", domain.RoleConductor)

	artists := []domain.Artist{composer, ensemble, conductor}
	track, _ := domain.NewTrack(1, 1, "Symphony No. 5", artists)
	album, _ := domain.NewAlbum(albumTitle, 1963)
	album.AddTrack(track)
	return album
}
