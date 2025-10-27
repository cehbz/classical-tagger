package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_AlbumArtistTag(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Actual   *domain.Album
		WantPass bool
		WantInfo int
	}{
		{
			Name: "pass - single track",
			Actual: &domain.Album{
				Title:        "Album",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Pollini", Role: domain.RoleSoloist},
						},
					},
				},
			},
			WantPass: true,
		},
		{
			Name: "info - dominant performer",
			Actual: &domain.Album{
				Title:        "Album",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name: "info - various composers",
			Actual: &domain.Album{
				Title:        "Various Artists Album",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name: "pass - no dominant performer",
			Actual: &domain.Album{
				Title:        "Album",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Pollini", Role: domain.RoleSoloist},
						},
					},
					{
						Disc:  1,
						Track: 2,
						Title: "Work 2",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Arrau", Role: domain.RoleSoloist},
						},
					},
					{
						Disc:  1,
						Track: 3,
						Title: "Work 3",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Brendel", Role: domain.RoleSoloist},
						},
					},
					{
						Disc:  1,
						Track: 4,
						Title: "Work 4",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Ashkenazy", Role: domain.RoleSoloist},
						},
					},
				},
			},
			WantPass: true,
		},
		{
			Name: "pass - two composers only",
			Actual: &domain.Album{
				Title:        "Album",
				OriginalYear: 1963,
				Tracks: []*domain.Track{
					{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
						},
					},
					{
						Disc:  1,
						Track: 2,
						Title: "Work 2",
						Artists: []domain.Artist{
							domain.Artist{Name: "Mozart", Role: domain.RoleComposer},
							domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.AlbumArtistTag(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				infoCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelInfo {
						infoCount++
					}
				}

				if infoCount != tt.WantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}
