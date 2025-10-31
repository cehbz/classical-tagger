package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_RecordLabelPresent(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		WantPass     bool
		WantWarnings int
	}{
		{
			Name:     "valid - both label and catalog present",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			WantPass: true,
		},
		{
			Name:         "warning - no edition at all",
			Actual:       NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("", "", 2010).Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - missing label",
			Actual:       NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("", "4776516", 2010).Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - missing catalog",
			Actual:       NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "", 2010).Build(),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - both missing",
			Actual:       NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("", "", 2010).Build(),
			WantPass:     false,
			WantWarnings: 1, // One warning for missing edition info
		},
		{
			Name:     "valid - Harmonia Mundi",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("harmonia mundi", "HMC902170", 2010).Build(),
			WantPass: true,
		},
		{
			Name:     "valid - Naxos",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Naxos", "8.557308", 2010).Build(),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.RecordLabelPresent(tt.Actual, nil)

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

func TestRules_RecordLabelAccuracy(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		Reference  *domain.Album
		WantPass   bool
		WantErrors int
	}{
		{
			Name:      "valid - exact match",
			Actual:    NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			Reference: NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			WantPass:  true,
		},
		{
			Name:       "error - label mismatch",
			Actual:     NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Sony Classical", "4776516", 2010).Build(),
			Reference:  NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - catalog mismatch",
			Actual:     NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "12345", 2010).Build(),
			Reference:  NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - both mismatch",
			Actual:     NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Sony Classical", "12345", 2010).Build(),
			Reference:  NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			WantPass:   false,
			WantErrors: 2,
		},
		{
			Name:      "pass - no reference edition",
			Actual:    NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			Reference: NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithoutEdition().Build(),
			WantPass:  true, // Can't validate without reference
		},
		{
			Name:      "pass - no actual edition but no reference either",
			Actual:    NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithoutEdition().Build(),
			Reference: NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithoutEdition().Build(),
			WantPass:  true,
		},
		{
			Name:      "pass - actual missing but no reference to check against",
			Actual:    NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithoutEdition().Build(),
			Reference: NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}).Build().WithEdition("Deutsche Grammophon", "4776516", 2010).Build(),
			WantPass:  true, // Presence checked by RecordLabelPresent, not accuracy
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.RecordLabelAccuracy(tt.Actual, tt.Reference)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelError {
						errorCount++
					}
				}

				if errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}
