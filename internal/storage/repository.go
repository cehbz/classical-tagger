package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// Repository handles JSON serialization and deserialization of torrents.
// No DTOs needed - domain objects serialize directly with JSON tags.
type Repository struct{}

// NewRepository creates a new Repository.
func NewRepository() *Repository {
	return &Repository{}
}

// SaveToJSON serializes a torrent to JSON bytes.
// Domain objects have JSON tags, so no DTO conversion needed.
func (r *Repository) SaveToJSON(torrent *domain.Torrent) ([]byte, error) {
	data, err := json.MarshalIndent(torrent, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal torrent: %w", err)
	}
	return data, nil
}

// LoadFromJSON deserializes a torrent from JSON bytes.
// Domain objects have JSON tags, so no DTO conversion needed.
func (r *Repository) LoadFromJSON(data []byte) (*domain.Torrent, error) {
	var torrent domain.Torrent
	if err := json.Unmarshal(data, &torrent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &torrent, nil
}

// SaveToFile saves a torrent to a JSON file.
func (r *Repository) SaveToFile(torrent *domain.Torrent, path string) error {
	data, err := r.SaveToJSON(torrent)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadFromFile loads a torrent from a JSON file.
func (r *Repository) LoadFromFile(path string) (*domain.Torrent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return r.LoadFromJSON(data)
}
