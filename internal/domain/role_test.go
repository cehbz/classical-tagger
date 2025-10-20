package domain

import "testing"

func TestRole_String(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want string
	}{
		{"composer", RoleComposer, "composer"},
		{"soloist", RoleSoloist, "soloist"},
		{"ensemble", RoleEnsemble, "ensemble"},
		{"conductor", RoleConductor, "conductor"},
		{"arranger", RoleArranger, "arranger"},
		{"guest", RoleGuest, "guest"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Role
		wantErr bool
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
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseRole() = %v, want %v", got, tt.want)
			}
		})
	}
}
