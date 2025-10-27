package domain

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		Name  string
		Level Level
		Want  string
	}{
		{"error level", LevelError, "ERROR"},
		{"warning level", LevelWarning, "WARNING"},
		{"info level", LevelInfo, "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := tt.Level.String(); got != tt.Want {
				t.Errorf("Level.String() = %v, want %v", got, tt.Want)
			}
		})
	}
}
