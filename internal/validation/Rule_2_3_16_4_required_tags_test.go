package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_RequiredTags(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		WantPass     bool
		WantErrors   int
		WantWarnings int
	}{
		{
			Name: "valid - all required tags present",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Symphony No. 5",
						Artists: []domain.Artist{
							{Name: "Ludwig van Beethoven", Role: domain.RoleComposer},
							{Name: "Vienna Phil", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass:     true,
			WantErrors:   0,
			WantWarnings: 0,
		},
		{
			Name: "missing year - warning only",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 0,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Symphony",
						Artists: []domain.Artist{
							{Name: "Beethoven", Role: domain.RoleComposer},
							{Name: "Vienna Phil", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass:     false,
			WantErrors:   0,
			WantWarnings: 1,
		},
		{
			Name: "missing track title",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "",
						Artists: []domain.Artist{
							{Name: "Beethoven", Role: domain.RoleComposer},
							{Name: "Vienna Phil", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "missing artists",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:    1,
						Track:   1,
						Title:   "Symphony",
						Artists: []domain.Artist{},
					},
				},
			},
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "only composer, no performers",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:    1,
						Track:   1,
						Title:   "Symphony",
						Artists: []domain.Artist{domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}},
					},
				},
			},
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "multiple tracks, some missing title",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:    1,
						Track:   1,
						Title:   "Symphony No. 1",
						Artists: []domain.Artist{domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}},
					},
					{
						Disc:    1,
						Track:   2,
						Title:   "",
						Artists: []domain.Artist{domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}},
					},
					{
						Disc:    1,
						Track:   3,
						Title:   "Symphony No. 3",
						Artists: []domain.Artist{domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}},
					},
				},
			},
			WantPass:   false,
			WantErrors: 1,
		},
		{
			Name: "multiple issues",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:    1,
						Track:   1,
						Title:   "",
						Artists: []domain.Artist{domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}},
					},
					{
						Disc:    1,
						Track:   2,
						Title:   "Symphony",
						Artists: []domain.Artist{},
					},
				},
			},
			WantPass:     false,
			WantErrors:   3, // Album title, track1 title, track2 artists
			WantWarnings: 1, // Year
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.RequiredTags(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelError {
						errorCount++
					}
					if issue.Level == domain.LevelWarning {
						warningCount++
					}
				}

				if errorCount != tt.WantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.WantErrors)
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
