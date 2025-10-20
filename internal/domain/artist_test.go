package domain

import "testing"

func TestNewArtist(t *testing.T) {
	tests := []struct {
		name    string
		argName string
		role    Role
		wantErr bool
	}{
		{
			name:    "valid artist",
			argName: "Hans-Christoph Rademann",
			role:    RoleConductor,
			wantErr: false,
		},
		{
			name:    "valid composer",
			argName: "Johann Sebastian Bach",
			role:    RoleComposer,
			wantErr: false,
		},
		{
			name:    "empty name",
			argName: "",
			role:    RoleConductor,
			wantErr: true,
		},
		{
			name:    "whitespace only name",
			argName: "   ",
			role:    RoleSoloist,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewArtist(tt.argName, tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewArtist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Name() != tt.argName {
					t.Errorf("NewArtist().Name() = %v, want %v", got.Name(), tt.argName)
				}
				if got.Role() != tt.role {
					t.Errorf("NewArtist().Role() = %v, want %v", got.Role(), tt.role)
				}
			}
		})
	}
}

func TestArtist_Name(t *testing.T) {
	artist, _ := NewArtist("RIAS Kammerchor Berlin", RoleEnsemble)
	
	if got := artist.Name(); got != "RIAS Kammerchor Berlin" {
		t.Errorf("Artist.Name() = %v, want %v", got, "RIAS Kammerchor Berlin")
	}
}

func TestArtist_Role(t *testing.T) {
	artist, _ := NewArtist("Felix Mendelssohn Bartholdy", RoleComposer)
	
	if got := artist.Role(); got != RoleComposer {
		t.Errorf("Artist.Role() = %v, want %v", got, RoleComposer)
	}
}
