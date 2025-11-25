package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Discogs struct {
		Token string `yaml:"token"`
	} `yaml:"discogs"`
	Redacted struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"redacted"`
	Cache struct {
		TTLHours int `yaml:"ttl_hours"` // Default: 24 if not specified
	} `yaml:"cache"`
}

// LoadDiscogsToken loads the Discogs personal access token from the config file.
func LoadDiscogsToken() (string, error) {
	configPath := getConfigPath()

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found at %s: please create it with your Discogs token", configPath)
		}
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check if token exists
	if cfg.Discogs.Token == "" {
		return "", fmt.Errorf("Discogs token not found in config file: please add 'discogs.token' to %s", configPath)
	}

	return cfg.Discogs.Token, nil
}

// LoadRedactedAPIKey loads the Redacted API key from the config file.
func LoadRedactedAPIKey() (string, error) {
	configPath := getConfigPath()

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found at %s: please create it with your Redacted API key", configPath)
		}
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check if API key exists
	if cfg.Redacted.APIKey == "" {
		return "", fmt.Errorf("redacted API key not found in config file: please add 'redacted.api_key' to %s", configPath)
	}

	return cfg.Redacted.APIKey, nil
}

// LoadCacheTTL loads the cache TTL from config file, returns default if not specified.
func LoadCacheTTL() time.Duration {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return 24 * time.Hour // Default
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return 24 * time.Hour // Default
	}

	if cfg.Cache.TTLHours <= 0 {
		return 24 * time.Hour // Default
	}

	return time.Duration(cfg.Cache.TTLHours) * time.Hour
}

// getConfigPath returns the path to the config file.
// Respects XDG Base Directory specification.
func getConfigPath() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "classical-tagger", "config.yaml")
	}

	// Fall back to ~/.config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME env var
		homeDir = os.Getenv("HOME")
	}

	return filepath.Join(homeDir, ".config", "classical-tagger", "config.yaml")
}

// GetConfigPathForDisplay returns the config file path for user display.
// This is a public wrapper around getConfigPath().
func GetConfigPathForDisplay() string {
	return getConfigPath()
}

// CreateSampleConfig creates a sample config file at the appropriate location.
func CreateSampleConfig() error {
	configPath := getConfigPath()

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Sample configuration
	sample := `# Classical Tagger Configuration

# Discogs API Settings
discogs:
  # Your personal access token from https://www.discogs.com/settings/developers
  token: "your-discogs-token-here"

# Redacted API Settings
redacted:
  # Your API key from Redacted user settings
  # Generate at: https://redacted.sh/user.php?action=edit (Access Settings)
  api_key: "your-redacted-api-key-here"

# Cache Settings (optional)
cache:
  # Cache TTL in hours (default: 24)
  ttl_hours: 24
`

	// Write sample config
	if err := os.WriteFile(configPath, []byte(sample), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Sample config created at: %s\n", configPath)
	fmt.Println("Please edit it and add your Discogs personal access token.")
	return nil
}
