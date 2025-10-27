package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// Repository handles JSON serialization and deserialization of albums.
// No DTOs needed - domain objects serialize directly with JSON tags.
type Repository struct{}

// NewRepository creates a new Repository.
func NewRepository() *Repository {
	return &Repository{}
}

// SaveToJSON serializes an album to JSON bytes.
// Domain objects have JSON tags, so no DTO conversion needed.
func (r *Repository) SaveToJSON(album *domain.Album) ([]byte, error) {
	data, err := json.MarshalIndent(album, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal album: %w", err)
	}
	return data, nil
}

// LoadFromJSON deserializes an album from JSON bytes.
// Domain objects have JSON tags, so no DTO conversion needed.
func (r *Repository) LoadFromJSON(data []byte) (*domain.Album, error) {
	var album domain.Album
	if err := json.Unmarshal(data, &album); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &album, nil
}

// SaveToFile saves an album to a JSON file.
func (r *Repository) SaveToFile(album *domain.Album, path string) error {
	data, err := r.SaveToJSON(album)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadFromFile loads an album from a JSON file.
func (r *Repository) LoadFromFile(path string) (*domain.Album, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return r.LoadFromJSON(data)
}
