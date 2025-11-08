package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoCombinedTags(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Torrent
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:     "valid - normal title",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Maurizio Pollini", Role: domain.RoleSoloist}, domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - comma-separated artists in ARTIST tag (allowed)",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("Pollini, Arrau, Orchestra", domain.RoleSoloist).Build().Build(),
			WantPass: true, // Multiple artists in single tag is now allowed
		},
		{
			Name:     "valid - ensemble name with 'and'",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Work").ClearArtists().WithArtist("London Symphony Orchestra and Chorus", domain.RoleEnsemble).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "info - combined works in title",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Symphony No. 1 / Symphony No. 2").Build().Build(),
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name:     "valid - movement subtitle with slash",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Allegro / Fast").Build().Build(),
			WantPass: true, // Short parts, not multiple works
		},
		{
			Name:         "warning - track number in title",
			Actual:       NewTorrent().ClearTracks().AddTrack().WithTrack(1).WithTitle("01 - Symphony No. 5").Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - track number prefix in title",
			Actual:       NewTorrent().ClearTracks().AddTrack().WithTrack(5).WithTitle("05. Allegro con brio").Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - track number with 'Track' prefix in title",
			Actual:       NewTorrent().ClearTracks().AddTrack().WithTrack(3).WithTitle("Track 3: Finale").Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:     "valid - number in title that's not a track number",
			Actual:   NewTorrent().ClearTracks().AddTrack().WithTitle("Symphony No. 5").Build().Build(),
			WantPass: true, // "No. 5" is part of the work title, not a track number
		},
		{
			Name:         "warning - disc number in album title without subtitle",
			Actual:       NewTorrent().WithTitle("Album Disc 1").ClearTracks().AddTrack().WithTitle("Track 1").Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:     "valid - disc number in album title with meaningful subtitle",
			Actual:   NewTorrent().WithTitle("The Fragile - Left Disc 1").ClearTracks().AddTrack().WithTitle("Track 1").Build().Build(),
			WantPass: true, // Has meaningful subtitle "Left"
		},
	}

	for _, tt := range tests {
		for _, track := range tt.Actual.Tracks() {
			t.Run(tt.Name, func(t *testing.T) {
				result := rules.NoCombinedTags(track, nil, tt.Actual, nil)

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
}
