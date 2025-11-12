package domain

import "testing"

func TestRole_String(t *testing.T) {
	tests := []struct {
		Name string
		Role Role
		Want string
	}{
		{"composer", RoleComposer, "composer"},
		{"soloist", RoleSoloist, "soloist"},
		{"ensemble", RoleEnsemble, "ensemble"},
		{"conductor", RoleConductor, "conductor"},
		{"arranger", RoleArranger, "arranger"},
		{"guest", RoleGuest, "guest"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := tt.Role.String(); got != tt.Want {
				t.Errorf("Role.String() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestRole_IsPerformer(t *testing.T) {
	tests := []struct {
		Name string
		Role Role
		Want bool
	}{
		{"composer", RoleComposer, false},
		{"soloist", RoleSoloist, true},
		{"ensemble", RoleEnsemble, true},
		{"conductor", RoleConductor, true},
		{"arranger", RoleArranger, false},
		{"guest", RoleGuest, true},
		{"unknown", RoleUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := tt.Role.IsPerformer(); got != tt.Want {
				t.Errorf("Role.IsPerformer() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		Name    string
		Input   string
		Want    Role
		WantErr bool
	}{
		{"valid composer", "composer", RoleComposer, false},
		{"valid soloist", "soloist", RoleSoloist, false},
		{"valid ensemble", "ensemble", RoleEnsemble, false},
		{"valid conductor", "conductor", RoleConductor, false},
		{"valid arranger", "arranger", RoleArranger, false},
		{"valid guest", "guest", RoleGuest, false},
		{"case insensitive", "COMPOSER", RoleComposer, false},
		{"mixed case", "Soloist", RoleSoloist, false},
		{"invalid role", "pianist", Role(0), true},
		{"empty string", "", Role(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := ParseRole(tt.Input)
			if (err != nil) != tt.WantErr {
				t.Errorf("ParseRole() error = %v, wantErr %v", err, tt.WantErr)
				return
			}
			if !tt.WantErr && got != tt.Want {
				t.Errorf("ParseRole() = %v, want %v", got, tt.Want)
			}
		})
	}
}
