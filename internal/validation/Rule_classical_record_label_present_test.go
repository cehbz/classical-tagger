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
			Actual:   buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			WantPass: true,
		},
		{
			Name:         "warning - no edition at all",
			Actual:       buildAlbumWithEdition("", ""),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - missing label",
			Actual:       buildAlbumWithEdition("", "4776516"),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - missing catalog",
			Actual:       buildAlbumWithEdition("Deutsche Grammophon", ""),
			WantPass:     false,
			WantWarnings: 1,
		},
		{
			Name:         "warning - both missing",
			Actual:       buildAlbumWithEdition("", ""),
			WantPass:     false,
			WantWarnings: 1, // One warning for missing edition info
		},
		{
			Name:     "valid - Harmonia Mundi",
			Actual:   buildAlbumWithEdition("harmonia mundi", "HMC902170"),
			WantPass: true,
		},
		{
			Name:     "valid - Naxos",
			Actual:   buildAlbumWithEdition("Naxos", "8.557308"),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.RecordLabelPresent(tt.Actual, tt.Actual)

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
			Actual:    buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			Reference: buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			WantPass:  true,
		},
		{
			Name:       "error - label mismatch",
			Actual:     buildAlbumWithEdition("Sony Classical", "4776516"),
			Reference:  buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - catalog mismatch",
			Actual:     buildAlbumWithEdition("Deutsche Grammophon", "12345"),
			Reference:  buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name:       "error - both mismatch",
			Actual:     buildAlbumWithEdition("Sony Classical", "12345"),
			Reference:  buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			WantPass:   false,
			WantErrors: 2,
		},
		{
			Name:      "pass - no reference edition",
			Actual:    buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			Reference: buildAlbumWithoutEdition(),
			WantPass:  true, // Can't validate without reference
		},
		{
			Name:      "pass - no actual edition but no reference either",
			Actual:    buildAlbumWithoutEdition(),
			Reference: buildAlbumWithoutEdition(),
			WantPass:  true,
		},
		{
			Name:      "pass - actual missing but no reference to check against",
			Actual:    buildAlbumWithoutEdition(),
			Reference: buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
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

// Helper to build album with edition information
func buildAlbumWithEdition(label, catalogNumber string) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}
	track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{composer, ensemble}}

	var edition *domain.Edition
	if label != "" || catalogNumber != "" {
		edition = &domain.Edition{Label: label, Year: 2010, CatalogNumber: catalogNumber}
	}

	return &domain.Album{
		Title:        "Beethoven Symphonies",
		OriginalYear: 1963,
		Edition:      edition,
		Tracks:       []*domain.Track{&track},
	}
}

// Helper to build album without edition
func buildAlbumWithoutEdition() *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Berlin Phil", Role: domain.RoleEnsemble}
	track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{composer, ensemble}}

	return &domain.Album{
		Title:        "Beethoven Symphonies",
		OriginalYear: 1963,
		Tracks:       []*domain.Track{&track},
	}
}
