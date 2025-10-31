package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerInFolderName(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		AlbumTitle   string
		ComposerName string
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:         "valid - full composer name in title",
			AlbumTitle:   "Ludwig van Beethoven - Symphony No. 5 [1963] [FLAC]",
			ComposerName: "Ludwig van Beethoven",
			WantPass:     true,
		},
		{
			Name:         "valid - abbreviated composer name",
			AlbumTitle:   "J.S. Bach - Brandenburg Concertos [1982] [FLAC]",
			ComposerName: "Johann Sebastian Bach",
			WantPass:     true,
		},
		{
			Name:         "info - last name only",
			AlbumTitle:   "Beethoven - Symphony No. 5 [1963] [FLAC]",
			ComposerName: "Ludwig van Beethoven",
			WantPass:     false,
			WantInfo:     1,
		},
		{
			Name:         "warning - composer missing",
			AlbumTitle:   "Symphony No. 5 [1963] [FLAC]",
			ComposerName: "Ludwig van Beethoven",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "valid - Various Artists",
			AlbumTitle:   "Various Artists - Classical Favorites [2000] [FLAC]",
			ComposerName: "Bach",
			WantPass:     true, // Various artists is an exception
		},
		{
			Name:         "valid - composer with particles",
			AlbumTitle:   "Beethoven - Piano Sonatas [1990] [FLAC]",
			ComposerName: "Ludwig van Beethoven",
			WantPass:     false,
			WantInfo:     1, // Has "Beethoven" but not full name
		},
		{
			Name:         "valid - full name present",
			AlbumTitle:   "Wolfgang Amadeus Mozart - Piano Concertos [1985] [FLAC]",
			ComposerName: "Wolfgang Amadeus Mozart",
			WantPass:     true,
		},
		{
			Name:         "valid - W.A. Mozart abbreviation",
			AlbumTitle:   "W.A. Mozart - Piano Concertos [1985] [FLAC]",
			ComposerName: "Wolfgang Amadeus Mozart",
			WantPass:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			actual := NewAlbum().WithTitle(tt.AlbumTitle).ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: tt.ComposerName, Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build()
			result := rules.ComposerInFolderName(actual, nil)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelWarning {
						warningCount++
					} else if issue.Level == domain.LevelInfo {
						infoCount++
					}
				}

				if tt.WantWarnings > 0 && warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}
				if tt.WantInfo > 0 && infoCount != tt.WantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}
