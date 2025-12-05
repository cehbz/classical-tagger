package domain

import "testing"

func TestParseArtistField(t *testing.T) {
	tests := []struct {
		Name        string
		Field       string
		WantArtists []Artist
	}{
		{
			Name:  "semicolon separated",
			Field: "Pollini; Berlin Phil; Karajan",
			WantArtists: []Artist{
				{Name: "Pollini", Role: RoleUnknown},
				{Name: "Berlin Phil", Role: RoleUnknown},
				{Name: "Karajan", Role: RoleUnknown},
			},
		},
		{
			Name:  "comma separated",
			Field: "Pollini, Berlin Philharmonic, Karajan",
			WantArtists: []Artist{
				{Name: "Pollini", Role: RoleUnknown},
				{Name: "Berlin Philharmonic", Role: RoleUnknown},
				{Name: "Karajan", Role: RoleUnknown},
			},
		},
		{
			Name:  "single artist",
			Field: "Maurizio Pollini",
			WantArtists: []Artist{
				{Name: "Maurizio Pollini", Role: RoleUnknown},
			},
		},
		{
			Name:  "with ensemble",
			Field: "RIAS Kammerchor; Hans-Christoph Rademann",
			WantArtists: []Artist{
				{Name: "RIAS Kammerchor", Role: RoleUnknown},
				{Name: "Hans-Christoph Rademann", Role: RoleUnknown},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := ParseArtistField(tt.Field)

			if len(got) != len(tt.WantArtists) {
				t.Errorf("ParseArtistField(%q) returned %d artists, want %d", tt.Field, len(got), len(tt.WantArtists))
			}

			for i := range got {
				if got[i] != tt.WantArtists[i] {
					t.Errorf("Artist %d = %+v, want %+v", i, got[i], tt.WantArtists[i])
				}
			}
		})
	}
}
