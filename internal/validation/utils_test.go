package validation

import "testing"


func TestLastName(t *testing.T) {
	tests := []struct {
		Name         string
		ComposerName string
		Want         string
	}{
		{"simple name", "Johann Bach", "Bach"},
		{"full name", "Johann Sebastian Bach", "Bach"},
		{"with particle", "Ludwig van Beethoven", "van Beethoven"},
		{"with initials", "J.S. Bach", "Bach"},
		{"reversed format", "Beethoven, Ludwig van", "Beethoven"},
		{"compound particle", "Felix Mendelssohn Bartholdy", "Bartholdy"},
		{"single word", "Bach", "Bach"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := lastName(tt.ComposerName)
			if got != tt.Want {
				t.Errorf("extractLastNames() = %v, want %v", got, tt.Want)
				return
			}
		})
	}
}
