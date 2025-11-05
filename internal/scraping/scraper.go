package scraping

import (
	"errors"
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/storage"
)

var (
	// ErrExtractionFailed indicates the extraction process failed
	ErrExtractionFailed = errors.New("extraction failed")

	// ErrUnsupportedURL indicates the URL is not supported by any extractor
	ErrUnsupportedURL = errors.New("unsupported URL")
)

// Extractor defines the interface for extracting album metadata from websites.
type Extractor interface {
	// Name returns the human-readable name of this extractor
	Name() string

	// CanHandle returns true if this extractor can handle the given URL
	CanHandle(url string) bool

	// Extract extracts album metadata from the given URL
	// Returns domain.Album (which is just domain.Album)
	Extract(url string) (*ExtractionResult, error)
}

// Registry manages available extractors.
type Registry struct {
	extractors []Extractor
}

// NewRegistry creates a new extractor registry.
func NewRegistry() *Registry {
	return &Registry{
		extractors: make([]Extractor, 0),
	}
}

// Register adds an extractor to the registry.
func (r *Registry) Register(extractor Extractor) {
	r.extractors = append(r.extractors, extractor)
}

// Get returns the first extractor that can handle the given URL.
func (r *Registry) Get(url string) Extractor {
	for _, extractor := range r.extractors {
		if extractor.CanHandle(url) {
			return extractor
		}
	}
	return nil
}

// Extract extracts album data from a URL using the appropriate extractor.
func (r *Registry) Extract(url string) (*ExtractionResult, error) {
	extractor := r.Get(url)
	if extractor == nil {
		return nil, fmt.Errorf("%w: no extractor for %s", ErrUnsupportedURL, url)
	}

	return extractor.Extract(url)
}

// SaveToJSON saves extracted torrent data to JSON format.
// Converts Album to Torrent if needed (for backward compatibility).
func SaveToJSON(torrentData *domain.Torrent) ([]byte, error) {
	repo := storage.NewRepository()
	jsonData, err := repo.SaveToJSON(torrentData)
	if err != nil {
		return nil, fmt.Errorf("failed to save JSON: %w", err)
	}
	return jsonData, nil
}

// DefaultRegistry returns a registry with all built-in extractors.
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// Register built-in extractors here as they're implemented
	// registry.Register(NewHarmoniaMundiExtractor())
	// registry.Register(NewNaxosExtractor())
	// etc.

	return registry
}

// SynthesizeMissingEditionData fills in missing required edition data with placeholder values.
// Returns true if any data was synthesized.
// Note: This function works with Album internally (for scraper compatibility).
func SynthesizeMissingEditionData(data *domain.Album) bool {
	if data.Edition == nil {
		return false
	}

	synthesized := false

	// If we have a catalog number but no label, synthesize a placeholder label
	if data.Edition.CatalogNumber != "" && data.Edition.Label == "" {
		data.Edition.Label = "[Unknown Label]"
		synthesized = true
	}

	// If edition year is missing, try to use original year, or synthesize
	if data.Edition.Year == 0 {
		if data.OriginalYear > 0 {
			// Use original year as edition year (reasonable default)
			data.Edition.Year = data.OriginalYear
			synthesized = true
		} else {
			// No year at all - synthesize a placeholder
			data.Edition.Year = 1900 // Placeholder year - obviously fake
			synthesized = true
		}
	}

	return synthesized
}
