package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FolderNameFormat(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		RootPath   string
		AlbumYear    int
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:       "valid - full format with FLAC",
			RootPath: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			AlbumYear:  1963,
			WantPass:   true,
		},
		{
			Name:       "valid - full format with MP3",
			RootPath: "Bach - Brandenburg Concertos [1982] [MP3]",
			AlbumYear:  1982,
			WantPass:   true,
		},
		{
			Name:       "info - missing format indicator",
			RootPath: "Mozart - Piano Concertos [1990]",
			AlbumYear:  1990,
			WantPass:   false,
			WantInfo:   1,
		},
		{
			Name:         "warning - missing year",
			RootPath:   "Vivaldi - The Four Seasons [FLAC]",
			AlbumYear:    1980,
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - missing separator",
			RootPath:   "Beethoven Symphony No. 5 [1963] [FLAC]",
			AlbumYear:    1963,
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - year mismatch",
			RootPath:   "Bach - Cello Suites [1990] [FLAC]",
			AlbumYear:    1985,
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "multiple issues",
			RootPath:   "Beethoven Symphony No. 5",
			AlbumYear:    1963,
			WantPass:     false,
			WantWarnings: 2, // No separator, no year
		},
		{
			Name:       "valid - with extra info",
			RootPath: "Beethoven - Symphony No. 5 [1963] [FLAC] [24-96]",
			AlbumYear:  1963,
			WantPass:   true,
		},
		{
			Name:       "valid - various artist format",
			RootPath: "Various Artists - Classical Favorites [2000] [FLAC]",
			AlbumYear:  2000,
			WantPass:   true,
		},
		{
			Name:       "valid - WAV format",
			RootPath: "Mahler - Symphony No. 2 [1991] [WAV]",
			AlbumYear:  1991,
			WantPass:   true,
		},
		{
			Name:       "valid - FLAC with quality info",
			RootPath: "Beethoven - Symphony No. 5 [1963] [FLAC 96-24]",
			AlbumYear:  1963,
			WantPass:   true,
		},
		{
			Name:       "valid - MP3 with quality",
			RootPath: "Bach - Brandenburg Concertos [1982] [MP3 V0]",
			AlbumYear:  1982,
			WantPass:   true,
		},
		{
			Name:       "valid - ALAC format",
			RootPath: "Debussy - Préludes [1985] [ALAC]",
			AlbumYear:  1985,
			WantPass:   true,
		},
		{
			Name:       "valid - year without brackets",
			RootPath: "Noël! Christmas! Weihnachten! (RIAS-Kammerchor, Rademann) - 2013 [FLAC]",
			AlbumYear:  2013,
			WantPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			actual := &domain.Torrent{
				RootPath:     tt.RootPath, // Use RootPath (directory name) instead of Title
				OriginalYear: tt.AlbumYear,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Symphony",
						Artists: []domain.Artist{{Name: "Beethoven", Role: domain.RoleComposer}, {Name: "Orchestra", Role: domain.RoleEnsemble}},
					},
				},
			}

			result := rules.FolderNameFormat(actual, nil)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues {
					switch issue.Level {
					case domain.LevelWarning:
						warningCount++
					case domain.LevelInfo:
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
