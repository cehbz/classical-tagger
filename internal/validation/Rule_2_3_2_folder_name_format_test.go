package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FolderNameFormat(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name         string
		albumTitle   string
		albumYear    int
		wantPass     bool
		wantWarnings int
		wantInfo     int
	}{
		{
			name:       "valid - full format with FLAC",
			albumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			albumYear:  1963,
			wantPass:   true,
		},
		{
			name:       "valid - full format with MP3",
			albumTitle: "Bach - Brandenburg Concertos [1982] [MP3]",
			albumYear:  1982,
			wantPass:   true,
		},
		{
			name:         "info - missing format indicator",
			albumTitle:   "Mozart - Piano Concertos [1990]",
			albumYear:    1990,
			wantPass:     false,
			wantInfo:     1,
		},
		{
			name:         "warning - missing year",
			albumTitle:   "Vivaldi - The Four Seasons [FLAC]",
			albumYear:    1980,
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - missing separator",
			albumTitle:   "Beethoven Symphony No. 5 [1963] [FLAC]",
			albumYear:    1963,
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - year mismatch",
			albumTitle:   "Bach - Cello Suites [1990] [FLAC]",
			albumYear:    1985,
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "multiple issues",
			albumTitle:   "Beethoven Symphony No. 5",
			albumYear:    1963,
			wantPass:     false,
			wantWarnings: 2, // No separator, no year
		},
		{
			name:       "valid - with extra info",
			albumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC] [24-96]",
			albumYear:  1963,
			wantPass:   true,
		},
		{
			name:       "valid - various artist format",
			albumTitle: "Various Artists - Classical Favorites [2000] [FLAC]",
			albumYear:  2000,
			wantPass:   true,
		},
		{
			name:       "valid - WAV format",
			albumTitle: "Mahler - Symphony No. 2 [1991] [WAV]",
			albumYear:  1991,
			wantPass:   true,
		},
		{
			name:       "valid - ALAC format",
			albumTitle: "Debussy - Préludes [1985] [ALAC]",
			albumYear:  1985,
			wantPass:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
			ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
			track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})
			actual, _ := domain.NewAlbum(tt.albumTitle, tt.albumYear)
			actual.AddTrack(track)
		
			result := rules.FolderNameFormat(actual, actual)
			
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
