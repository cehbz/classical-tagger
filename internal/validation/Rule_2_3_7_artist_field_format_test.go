package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArtistFieldFormat(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Torrent
		Reference    *domain.Torrent
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name:     "valid - has performers",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:         "warning - only composer",
			Actual:       NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}).Build().Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:     "valid - just performers (no composer)",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:     "valid - ensemble only",
			Actual:   NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Emerson Quartet", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass: true,
		},
		{
			Name:      "info - performer count differs from reference",
			Actual:    NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}).Build().Build(),
			Reference: NewTorrent().WithTitle("Classical Album").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}, domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}).Build().Build(),
			WantPass:  false,
			WantInfo:  1,
		},
	}

	for _, tt := range tests {
		for i, track := range tt.Actual.Tracks() {
			t.Run(tt.Name, func(t *testing.T) {
				var refTrack *domain.Track
				if tt.Reference != nil {
					refTracks := tt.Reference.Tracks()
					if i < len(refTracks) {
						refTrack = refTracks[i]
					}
				}
				result := rules.ArtistFieldFormat(track, refTrack, tt.Actual, tt.Reference)
				if result.Passed() != tt.WantPass {
					t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
				}

				errorCount := 0
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues {
					switch issue.Level {
					case domain.LevelError:
						errorCount++
					case domain.LevelWarning:
						warningCount++
					case domain.LevelInfo:
						infoCount++
					}
				}

				dumpIssues := false
				if tt.WantPass && errorCount > 0 {
					t.Errorf("Errors = %d, want 0", errorCount)
					dumpIssues = true
				}
				if tt.WantWarnings != warningCount {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.WantWarnings)
					dumpIssues = true
				}
				if tt.WantInfo != infoCount {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
					dumpIssues = true
				}

				if dumpIssues {
					for _, issue := range result.Issues {
						t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
					}
				}
			})
		}
	}
}

func TestTrack_Performers(t *testing.T) {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	soloist := domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	arranger := domain.Artist{Name: "Mahler", Role: domain.RoleArranger}

	tests := []struct {
		Name  string
		Track domain.Track
		Want  []string
	}{
		{
			Name:  "all roles",
			Track: domain.Track{Artists: []domain.Artist{composer, soloist, ensemble, arranger}},
			Want:  []string{"Pollini", "Orchestra"},
		},
		{
			Name:  "only composer",
			Track: domain.Track{Artists: []domain.Artist{composer}},
			Want:  []string{},
		},
		{
			Name:  "only performers",
			Track: domain.Track{Artists: []domain.Artist{soloist, ensemble}},
			Want:  []string{"Pollini", "Orchestra"},
		},
		{
			Name:  "empty",
			Track: domain.Track{Artists: []domain.Artist{}},
			Want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := tt.Track.Performers()
			if len(got) != len(tt.Want) {
				t.Errorf("Track.Performers() count = %d, want %d", len(got), len(tt.Want))
				return
			}
			for i := range got {
				if got[i] != tt.Want[i] {
					t.Errorf("Track.Performers()[%d] = %q, want %q", i, got[i], tt.Want[i])
				}
			}
		})
	}
}
