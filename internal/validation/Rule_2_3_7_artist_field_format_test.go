package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArtistFieldFormat(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		Reference    *domain.Album
		WantPass     bool
		WantWarnings int
		WantInfo     int
	}{
		{
			Name: "valid - has performers",
			Actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Pollini", domain.RoleSoloist,
				"Berlin Phil", domain.RoleEnsemble,
			),
			WantPass: true,
		},
		{
			Name: "warning - only composer",
			Actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
			),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name: "valid - just performers (no composer)",
			Actual: buildAlbumWithArtists(
				"Pollini", domain.RoleSoloist,
				"Berlin Phil", domain.RoleEnsemble,
			),
			WantPass: true,
		},
		{
			Name: "valid - ensemble only",
			Actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Emerson Quartet", domain.RoleEnsemble,
			),
			WantPass: true,
		},
		{
			Name: "info - performer count differs from reference",
			Actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Pollini", domain.RoleSoloist,
			),
			Reference: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Pollini", domain.RoleSoloist,
				"Orchestra", domain.RoleEnsemble,
			),
			WantPass: false,
			WantInfo: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.ArtistFieldFormat(tt.Actual, tt.Reference)

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

func TestGetPerformers(t *testing.T) {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	soloist := domain.Artist{Name: "Pollini", Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}
	arranger := domain.Artist{Name: "Mahler", Role: domain.RoleArranger}

	tests := []struct {
		Name    string
		Artists []domain.Artist
		Want    []string
	}{
		{
			Name:    "all roles",
			Artists: []domain.Artist{composer, soloist, ensemble, arranger},
			Want:    []string{"Pollini", "Orchestra"},
		},
		{
			Name:    "only composer",
			Artists: []domain.Artist{composer},
			Want:    []string{},
		},
		{
			Name:    "only performers",
			Artists: []domain.Artist{soloist, ensemble},
			Want:    []string{"Pollini", "Orchestra"},
		},
		{
			Name:    "empty",
			Artists: []domain.Artist{},
			Want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := getPerformers(tt.Artists)
			if len(got) != len(tt.Want) {
				t.Errorf("getPerformers() count = %d, want %d", len(got), len(tt.Want))
				return
			}
			for i := range got {
				if got[i] != tt.Want[i] {
					t.Errorf("getPerformers()[%d] = %q, want %q", i, got[i], tt.Want[i])
				}
			}
		})
	}
}
