package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoDiscNumbersInAlbumTag(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name         string
		albumTitle   string
		wantPass     bool
		wantWarnings int
	}{
		{
			name:       "valid - no disc numbers",
			albumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			wantPass:   true,
		},
		{
			name:         "warning - disc number in title",
			albumTitle:   "Beethoven - Symphony No. 5 - Disc 1 [1963] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - CD number in title",
			albumTitle:   "Bach - Brandenburg Concertos CD 2 [1982] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - disk spelling",
			albumTitle:   "Mozart - Symphonies Disk 1 [1990] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "valid - legitimate volume series",
			albumTitle: "Complete Works, Volume 1 [1995] [FLAC]",
			wantPass:   true, // Legitimate series name
		},
		{
			name:       "valid - collected recordings volume",
			albumTitle: "Collected Recordings, Vol. 2 [2000] [FLAC]",
			wantPass:   true,
		},
		{
			name:       "valid - anthology volume",
			albumTitle: "Classical Anthology, Volume 3 [2005] [FLAC]",
			wantPass:   true,
		},
		{
			name:       "valid - volume range",
			albumTitle: "Complete Symphonies, Volumes 1-5 [1998] [FLAC]",
			wantPass:   true, // Range indicates series
		},
		{
			name:       "valid - edition volume",
			albumTitle: "The Decca Edition, Volume 12 [2010] [FLAC]",
			wantPass:   true,
		},
		{
			name:         "warning - case insensitive",
			albumTitle:   "Symphony No. 5 DISC 1 [1963] [FLAC]",
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:       "valid - CD in catalog number",
			albumTitle: "Beethoven - Symphonies [BIS-CD-123] [1990] [FLAC]",
			wantPass:   true, // CD in catalog, not disc indicator
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildAlbumWithTitle(tt.albumTitle, "1963")
			result := rules.NoDiscNumbersInAlbumTag(actual, actual)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass {
				warningCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelWarning {
						warningCount++
					}
				}
				
				if warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}
				
				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

func TestIsLegitimateVolumeTitle(t *testing.T) {
	tests := []struct {
		title string
		want  bool
	}{
		{"Complete Works, Volume 1", true},
		{"Collected Recordings, Vol. 2", true},
		{"Classical Anthology, Volume 3", true},
		{"The Decca Edition, Volume 12", true},
		{"Symphonies, Volumes 1-5", true},
		{"Symphony No. 5 - Disc 1", false},
		{"Piano Concertos CD 2", false},
		{"String Quartets Disk 3", false},
		{"Bach Series, Vol. 1", true},
		{"Complete Collection, Volume 1-10", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := isLegitimateVolumeTitle(tt.title)
			if got != tt.want {
				t.Errorf("isLegitimateVolumeTitle(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}
