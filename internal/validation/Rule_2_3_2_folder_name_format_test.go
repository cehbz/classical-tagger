package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FolderNameFormat(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		AlbumTitle   string
		AlbumYear    int
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:       "valid - full format with FLAC",
			AlbumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			AlbumYear:  1963,
			WantPass:   true,
		},
		{
			Name:       "valid - full format with MP3",
			AlbumTitle: "Bach - Brandenburg Concertos [1982] [MP3]",
			AlbumYear:  1982,
			WantPass:   true,
		},
		{
			Name:       "info - missing format indicator",
			AlbumTitle: "Mozart - Piano Concertos [1990]",
			AlbumYear:  1990,
			WantPass:   false,
			WantInfo:   1,
		},
		{
			Name:         "warning - missing year",
			AlbumTitle:   "Vivaldi - The Four Seasons [FLAC]",
			AlbumYear:    1980,
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - missing separator",
			AlbumTitle:   "Beethoven Symphony No. 5 [1963] [FLAC]",
			AlbumYear:    1963,
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - year mismatch",
			AlbumTitle:   "Bach - Cello Suites [1990] [FLAC]",
			AlbumYear:    1985,
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "multiple issues",
			AlbumTitle:   "Beethoven Symphony No. 5",
			AlbumYear:    1963,
			WantPass:     false,
			WantWarnings: 2, // No separator, no year
		},
		{
			Name:       "valid - with extra info",
			AlbumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC] [24-96]",
			AlbumYear:  1963,
			WantPass:   true,
		},
		{
			Name:       "valid - various artist format",
			AlbumTitle: "Various Artists - Classical Favorites [2000] [FLAC]",
			AlbumYear:  2000,
			WantPass:   true,
		},
		{
			Name:       "valid - WAV format",
			AlbumTitle: "Mahler - Symphony No. 2 [1991] [WAV]",
			AlbumYear:  1991,
			WantPass:   true,
		},
		{
			Name:       "valid - ALAC format",
			AlbumTitle: "Debussy - Préludes [1985] [ALAC]",
			AlbumYear:  1985,
			WantPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
			ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
			track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{composer, ensemble}}
			actual := domain.Album{Title: tt.AlbumTitle, OriginalYear: tt.AlbumYear}
			actual.Tracks = append(actual.Tracks, &track)

			result := rules.FolderNameFormat(&actual, &actual)

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
