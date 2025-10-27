package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoDiscNumbersInAlbumTag(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		AlbumTitle   string
		WantPass     bool
		WantWarnings int
	}{
		{
			Name:       "valid - no disc numbers",
			AlbumTitle: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			WantPass:   true,
		},
		{
			Name:         "warning - disc number in title",
			AlbumTitle:   "Beethoven - Symphony No. 5 - Disc 1 [1963] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - CD number in title",
			AlbumTitle:   "Bach - Brandenburg Concertos CD 2 [1982] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - disk spelling",
			AlbumTitle:   "Mozart - Symphonies Disk 1 [1990] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "valid - legitimate volume series",
			AlbumTitle: "Complete Works, Volume 1 [1995] [FLAC]",
			WantPass:   true, // Legitimate series name
		},
		{
			Name:       "valid - collected recordings volume",
			AlbumTitle: "Collected Recordings, Vol. 2 [2000] [FLAC]",
			WantPass:   true,
		},
		{
			Name:       "valid - anthology volume",
			AlbumTitle: "Classical Anthology, Volume 3 [2005] [FLAC]",
			WantPass:   true,
		},
		{
			Name:       "valid - volume range",
			AlbumTitle: "Complete Symphonies, Volumes 1-5 [1998] [FLAC]",
			WantPass:   true, // Range indicates series
		},
		{
			Name:       "valid - edition volume",
			AlbumTitle: "The Decca Edition, Volume 12 [2010] [FLAC]",
			WantPass:   true,
		},
		{
			Name:         "warning - case insensitive",
			AlbumTitle:   "Symphony No. 5 DISC 1 [1963] [FLAC]",
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:       "valid - CD in catalog number",
			AlbumTitle: "Beethoven - Symphonies [BIS-CD-123] [1990] [FLAC]",
			WantPass:   true, // CD in catalog, not disc indicator
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			actual := buildAlbumWithTitle(tt.AlbumTitle, "1963")
			result := rules.NoDiscNumbersInAlbumTag(actual, actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				warningCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelWarning {
						warningCount++
					}
				}

				if warningCount != tt.WantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
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
