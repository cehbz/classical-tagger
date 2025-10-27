package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoLeadingSpaces(t *testing.T) {
	rules := NewRules()
	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name: "valid - no leading spaces",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Symphony No. 1",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: "01 - Symphony.flac",
					},
					{
						Disc:  1,
						Track: 2,
						Title: "Symphony No. 2",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: "02 - Concerto.flac",
					},
				},
			},
			WantPass: true,
		},
		{
			Name: "album title with leading space",
			Actual: &domain.Album{
				Title:        " Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Symphony",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: "01 - Symphony.flac",
					},
				},
			},
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "filename with leading space",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Symphony",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: " 01 - Symphony.flac",
					},
				},
			},
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "track title with leading space",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc: 1, Track: 1, Title: " Symphony No. 1", Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: "01 - Symphony.flac",
					},
				},
			},
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "folder name in path with leading space",
			Actual: &domain.Album{
				Title:        "Beethoven Symphonies",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Symphony",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: " CD1/01 - Symphony.flac",
					},
				},
			},
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "multiple violations",
			Actual: &domain.Album{
				Title:        " Beethoven",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: " Symphony No. 1",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: " 01 - Symphony.flac",
					},
					{
						Disc:  1,
						Track: 2,
						Title: "Concerto",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: "02 - Concerto.flac",
					},
				},
			},
			WantPass:   false,
			WantIssues: 3, // Album title + track1 filename + track1 title
		},
		{
			Name: "valid multi-disc with subfolders",
			Actual: &domain.Album{
				Title:        "Beethoven",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Symphony",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: "CD1/01 - Symphony.flac",
					},
					{
						Disc:  2,
						Track: 1,
						Title: "Concerto",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
							{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}},
						Name: "CD2/01 - Concerto.flac",
					},
				},
			},
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.NoLeadingSpaces(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass && len(result.Issues) != tt.WantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues), tt.WantIssues)
				for _, issue := range result.Issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}
