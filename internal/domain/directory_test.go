package domain

import (
	"testing"
)

func TestGenerateDirectoryName(t *testing.T) {
	tests := []struct {
		Name            string
		Torrent         *Torrent
		Want            string // We'll check that it contains expected parts rather than exact match
		WantContains    []string
		WantNotContains []string
	}{
		{
			Name: "full format with composer, performers, and year",
			Torrent: &Torrent{
				Title:        "Goldberg Variations",
				OriginalYear: 1981,
				Files: []FileLike{
					&Track{
						Track: 1,
						Title: "Aria",
						Artists: []Artist{
							{Name: "Johann Sebastian Bach", Role: RoleComposer},
							{Name: "Glenn Gould", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 2,
						Title: "Variation 1",
						Artists: []Artist{
							{Name: "Johann Sebastian Bach", Role: RoleComposer},
							{Name: "Glenn Gould", Role: RoleSoloist},
						},
					},
				},
			},
			WantContains: []string{"Bach", "Goldberg Variations", "1981", "[FLAC]"},
		},
		{
			Name: "composer and year without performers",
			Torrent: &Torrent{
				Title:        "Symphony No. 5",
				OriginalYear: 1970,
				Files: []FileLike{
					&Track{
						Track: 1,
						Title: "Allegro",
						Artists: []Artist{
							{Name: "Ludwig van Beethoven", Role: RoleComposer},
						},
					},
				},
			},
			WantContains:    []string{"Beethoven", "Symphony No. 5", "1970", "[FLAC]"},
			WantNotContains: []string{"("}, // No performers, so no parentheses
		},
		{
			Name: "no composer",
			Torrent: &Torrent{
				Title:        "Unknown Album",
				OriginalYear: 2000,
				Files: []FileLike{
					&Track{
						Track: 1,
						Title: "Track 1",
						Artists: []Artist{
							{Name: "Performer", Role: RoleSoloist},
						},
					},
				},
			},
			WantContains: []string{"Unknown Album", "[FLAC]"},
		},
		{
			Name: "no year",
			Torrent: &Torrent{
				Title: "Album Without Year",
				Files: []FileLike{
					&Track{
						Track: 1,
						Title: "Track 1",
						Artists: []Artist{
							{Name: "Composer", Role: RoleComposer},
						},
					},
				},
			},
			WantContains: []string{"Composer", "Album Without Year", "[FLAC]"},
		},
		{
			Name: "empty torrent",
			Torrent: &Torrent{
				Title: "Empty Album",
				Files: []FileLike{},
			},
			WantContains: []string{"Empty Album", "[FLAC]"},
		},
		{
			Name: "multiple composers",
			Torrent: &Torrent{
				Title:        "Mixed Composers",
				OriginalYear: 1990,
				Files: []FileLike{
					&Track{
						Track: 1,
						Title: "Track 1",
						Artists: []Artist{
							{Name: "Johann Sebastian Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 2,
						Title: "Track 2",
						Artists: []Artist{
							{Name: "Wolfgang Amadeus Mozart", Role: RoleComposer},
						},
					},
				},
			},
			WantContains: []string{"Mixed Composers", "[FLAC]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := GenerateDirectoryName(tt.Torrent)

			// Check length constraint (rule 2.3.12)
			if len(got) > 180 {
				t.Errorf("GenerateDirectoryName() length = %d, want <= 180", len(got))
			}

			// Check contains expected parts
			for _, want := range tt.WantContains {
				if !contains(got, want) {
					t.Errorf("GenerateDirectoryName() = %q, want to contain %q", got, want)
				}
			}

			// Check doesn't contain unexpected parts
			for _, notWant := range tt.WantNotContains {
				if contains(got, notWant) {
					t.Errorf("GenerateDirectoryName() = %q, should not contain %q", got, notWant)
				}
			}

			// Ensure it's not empty
			if got == "" {
				t.Error("GenerateDirectoryName() returned empty string")
			}
		})
	}
}

func TestSanitizeDirectoryName(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "normal name",
			Input: "Album Name",
			Want:  "Album Name",
		},
		{
			Name:  "invalid characters",
			Input: "Album: \"Name\" / Path\\Dir",
			Want:  "Album Name PathDir",
		},
		{
			Name:  "leading/trailing spaces",
			Input: "  Album Name  ",
			Want:  "Album Name",
		},
		{
			Name:  "Windows reserved name",
			Input: "CON",
			Want:  "_CON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := SanitizeDirectoryName(tt.Input)
			if got != tt.Want {
				t.Errorf("SanitizeDirectoryName() = %q, want %q", got, tt.Want)
			}
		})
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
