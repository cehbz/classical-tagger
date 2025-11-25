package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDiscogsToken(t *testing.T) {
	// Create temp config directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "classical-tagger")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}
	configFile := filepath.Join(configDir, "config.yaml")

	// Test case 1: Valid config file
	configContent := `discogs:
  token: "test-token-123"`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Override config path for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Also support XDG_CONFIG_HOME
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	token, err := LoadDiscogsToken()
	if err != nil {
		t.Fatalf("LoadDiscogsToken() error = %v", err)
	}

	if token != "test-token-123" {
		t.Errorf("Expected token 'test-token-123', got %s", token)
	}
}

func TestLoadDiscogsToken_MissingFile(t *testing.T) {
	// Use non-existent directory
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	defer os.Unsetenv("XDG_CONFIG_HOME")

	token, err := LoadDiscogsToken()
	if err == nil {
		t.Error("Expected error for missing config file")
	}
	if token != "" {
		t.Errorf("Expected empty token, got %s", token)
	}
}

func TestLoadDiscogsToken_MissingToken(t *testing.T) {
	// Create temp config directory with config missing token
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `other:
  setting: "value"`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	token, err := LoadDiscogsToken()
	if err == nil {
		t.Error("Expected error for missing token in config")
	}
	if token != "" {
		t.Errorf("Expected empty token, got %s", token)
	}
}

func TestLoadRedactedAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "classical-tagger")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}
	configFile := filepath.Join(configDir, "config.yaml")

	configContent := `redacted:
  api_key: "test-redacted-key"`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	apiKey, err := LoadRedactedAPIKey()
	if err != nil {
		t.Fatalf("LoadRedactedAPIKey() error = %v", err)
	}

	if apiKey != "test-redacted-key" {
		t.Errorf("Expected API key 'test-redacted-key', got %s", apiKey)
	}
}

func TestLoadCacheTTL(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "classical-tagger")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}
	configFile := filepath.Join(configDir, "config.yaml")

	configContent := `cache:
  ttl_hours: 48`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	ttl := LoadCacheTTL()
	expected := 48 * time.Hour

	if ttl != expected {
		t.Errorf("Expected TTL %v, got %v", expected, ttl)
	}
}

func TestLoadCacheTTL_Default(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	defer os.Unsetenv("XDG_CONFIG_HOME")

	ttl := LoadCacheTTL()
	expected := 24 * time.Hour

	if ttl != expected {
		t.Errorf("Expected default TTL %v, got %v", expected, ttl)
	}
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		xdgHome  string
		home     string
		expected string
	}{
		{
			name:     "XDG_CONFIG_HOME set",
			xdgHome:  "/custom/config",
			home:     "/home/user",
			expected: "/custom/config/classical-tagger/config.yaml",
		},
		{
			name:     "Use HOME directory",
			xdgHome:  "",
			home:     "/home/user",
			expected: "/home/user/.config/classical-tagger/config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.xdgHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgHome)
				defer os.Unsetenv("XDG_CONFIG_HOME")
			}
			if tt.home != "" {
				oldHome := os.Getenv("HOME")
				os.Setenv("HOME", tt.home)
				defer os.Setenv("HOME", oldHome)
			}

			path := getConfigPath()
			if path != tt.expected {
				t.Errorf("getConfigPath() = %s, want %s", path, tt.expected)
			}
		})
	}
}
