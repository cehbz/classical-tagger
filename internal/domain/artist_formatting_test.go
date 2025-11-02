package domain

import "testing"

// TestFormatArtists tests formatting multiple artists according to classical music rules.
func TestFormatArtists(t *testing.T) {
	tests := []struct {
		Name    string
		Artists []Artist
		Want    string
	}{
		{
			Name:    "single soloist",
			Artists: []Artist{{Name: "Glenn Gould", Role: RoleSoloist}},
			Want:    "Glenn Gould",
		},
		{
			Name: "soloist and ensemble",
			Artists: []Artist{
				{Name: "Anne-Sophie Mutter", Role: RoleSoloist},
				{Name: "Berlin Philharmonic", Role: RoleEnsemble},
			},
			Want: "Anne-Sophie Mutter, Berlin Philharmonic",
		},
		{
			Name: "soloist, ensemble, and conductor",
			Artists: []Artist{
				{Name: "Daniel Barenboim", Role: RoleConductor},
				{Name: "Yo-Yo Ma", Role: RoleSoloist},
				{Name: "Chicago Symphony Orchestra", Role: RoleEnsemble},
			},
			Want: "Yo-Yo Ma, Chicago Symphony Orchestra, Daniel Barenboim",
		},
		{
			Name: "multiple soloists",
			Artists: []Artist{
				{Name: "Martha Argerich", Role: RoleSoloist},
				{Name: "Daniel Barenboim", Role: RoleSoloist},
			},
			Want: "Martha Argerich, Daniel Barenboim",
		},
		{
			Name: "just ensemble",
			Artists: []Artist{
				{Name: "The Academy of Ancient Music", Role: RoleEnsemble},
			},
			Want: "The Academy of Ancient Music",
		},
		{
			Name: "ensemble and conductor",
			Artists: []Artist{
				{Name: "Claudio Abbado", Role: RoleConductor},
				{Name: "London Symphony Orchestra", Role: RoleEnsemble},
			},
			Want: "London Symphony Orchestra, Claudio Abbado",
		},
		{
			Name:    "empty artists",
			Artists: []Artist{},
			Want:    "",
		},
		{
			Name: "composer excluded",
			Artists: []Artist{
				{Name: "Johann Sebastian Bach", Role: RoleComposer},
				{Name: "Glenn Gould", Role: RoleSoloist},
			},
			Want: "Glenn Gould",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := FormatArtists(tt.Artists)
			if got != tt.Want {
				t.Errorf("FormatArtists() = %q, want %q", got, tt.Want)
			}
		})
	}
}
