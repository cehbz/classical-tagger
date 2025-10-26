package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_RecordLabelPresent(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		actual       *domain.Album
		wantPass     bool
		wantWarnings int
	}{
		{
			name:     "valid - both label and catalog present",
			actual:   buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			wantPass: true,
		},
		{
			name:         "warning - no edition at all",
			actual:       buildAlbumWithEdition("", ""),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - missing label",
			actual:       buildAlbumWithEdition("", "4776516"),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - missing catalog",
			actual:       buildAlbumWithEdition("Deutsche Grammophon", ""),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name:         "warning - both missing",
			actual:       buildAlbumWithEdition("", ""),
			wantPass:     false,
			wantWarnings: 1, // One warning for missing edition info
		},
		{
			name:     "valid - Harmonia Mundi",
			actual:   buildAlbumWithEdition("harmonia mundi", "HMC902170"),
			wantPass: true,
		},
		{
			name:     "valid - Naxos",
			actual:   buildAlbumWithEdition("Naxos", "8.557308"),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.RecordLabelPresent(tt.actual, tt.actual)

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

func TestRules_RecordLabelAccuracy(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name       string
		actual     *domain.Album
		reference  *domain.Album
		wantPass   bool
		wantErrors int
	}{
		{
			name:      "valid - exact match",
			actual:    buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			reference: buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			wantPass:  true,
		},
		{
			name:       "error - label mismatch",
			actual:     buildAlbumWithEdition("Sony Classical", "4776516"),
			reference:  buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - catalog mismatch",
			actual:     buildAlbumWithEdition("Deutsche Grammophon", "12345"),
			reference:  buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name:       "error - both mismatch",
			actual:     buildAlbumWithEdition("Sony Classical", "12345"),
			reference:  buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			wantPass:   false,
			wantErrors: 2,
		},
		{
			name:      "pass - no reference edition",
			actual:    buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			reference: buildAlbumWithoutEdition(),
			wantPass:  true, // Can't validate without reference
		},
		{
			name:      "pass - no actual edition but no reference either",
			actual:    buildAlbumWithoutEdition(),
			reference: buildAlbumWithoutEdition(),
			wantPass:  true,
		},
		{
			name:      "pass - actual missing but no reference to check against",
			actual:    buildAlbumWithoutEdition(),
			reference: buildAlbumWithEdition("Deutsche Grammophon", "4776516"),
			wantPass:  true, // Presence checked by RecordLabelPresent, not accuracy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.RecordLabelAccuracy(tt.actual, tt.reference)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				errorCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelError {
						errorCount++
					}
				}

				if errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

// Helper to build album with edition information
func buildAlbumWithEdition(label, catalogNumber string) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})

	var edition domain.Edition
	if label != "" || catalogNumber != "" {
		edition, _ = domain.NewEdition(label, 2010)
		edition = edition.WithCatalogNumber(catalogNumber)
	}

	album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
	album.AddTrack(track)
	album.WithEdition(edition)
	return album
}

// Helper to build album without edition
func buildAlbumWithoutEdition() *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Berlin Phil", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Symphony", []domain.Artist{composer, ensemble})

	album, _ := domain.NewAlbum("Beethoven Symphonies", 1963)
	album.AddTrack(track)
	return album
}
