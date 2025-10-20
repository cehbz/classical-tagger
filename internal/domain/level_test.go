package domain

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{"error level", LevelError, "ERROR"},
		{"warning level", LevelWarning, "WARNING"},
		{"info level", LevelInfo, "INFO"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
