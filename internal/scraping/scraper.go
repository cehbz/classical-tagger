package scraping

import (
	"errors"
	"fmt"
	"strings"

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
	Extract(url string) (*ExtractionResult, error) // ← Returns result wrapper
}

// AlbumData represents extracted album metadata before domain conversion.
type AlbumData struct {
	Title        string
	OriginalYear int
	Edition      *EditionData
	Tracks       []TrackData
}

// EditionData represents edition information.
type EditionData struct {
	Label         string
	CatalogNumber string
	EditionYear   int
}

// TrackData represents a single track.
type TrackData struct {
	Disc     int
	Track    int
	Title    string
	Composer string
	Artists  []ArtistData
	Arranger string // Optional
}

// ArtistData represents an artist.
type ArtistData struct {
	Name string
	Role string // "soloist", "ensemble", "conductor", etc.
}

// ToAlbum converts AlbumData to a domain Album.
func (a *AlbumData) ToAlbum() (*domain.Album, error) {
	// Create album
	album, err := domain.NewAlbum(a.Title, a.OriginalYear)
	if err != nil {
		return nil, fmt.Errorf("failed to create album: %w", err)
	}

	// Add edition if present
	if a.Edition != nil {
		edition, err := domain.NewEdition(a.Edition.Label, a.Edition.EditionYear)
		if err != nil {
			return nil, fmt.Errorf("failed to create edition: %w", err)
		}
		if a.Edition.CatalogNumber != "" {
			edition = edition.WithCatalogNumber(a.Edition.CatalogNumber)
		}
		album = album.WithEdition(edition)
	}

	// Add tracks
	for _, trackData := range a.Tracks {
		track, err := trackData.ToTrack()
		if err != nil {
			return nil, fmt.Errorf("failed to create track %d: %w", trackData.Track, err)
		}
		err = album.AddTrack(track)
		if err != nil {
			return nil, fmt.Errorf("failed to add track %d: %w", trackData.Track, err)
		}
	}

	return album, nil
}

// ToTrack converts TrackData to a domain Track.
func (t *TrackData) ToTrack() (*domain.Track, error) {
	// Create composer
	composer, err := domain.NewArtist(t.Composer, domain.RoleComposer)
	if err != nil {
		return nil, fmt.Errorf("invalid composer: %w", err)
	}

	// Create artists list starting with composer
	artists := []domain.Artist{composer}

	// Add other artists
	for _, artistData := range t.Artists {
		role, err := parseRole(artistData.Role)
		if err != nil {
			return nil, fmt.Errorf("invalid role %q: %w", artistData.Role, err)
		}

		artist, err := domain.NewArtist(artistData.Name, role)
		if err != nil {
			return nil, fmt.Errorf("invalid artist %q: %w", artistData.Name, err)
		}

		artists = append(artists, artist)
	}

	// Create track
	track, err := domain.NewTrack(t.Disc, t.Track, t.Title, artists)
	if err != nil {
		return nil, fmt.Errorf("failed to create track: %w", err)
	}

	return track, nil
}

// parseRole converts a string role to a domain.Role.
func parseRole(role string) (domain.Role, error) {
	switch strings.ToLower(role) {
	case "composer":
		return domain.RoleComposer, nil
	case "soloist", "solo":
		return domain.RoleSoloist, nil
	case "ensemble", "orchestra", "choir", "quartet":
		return domain.RoleEnsemble, nil
	case "conductor":
		return domain.RoleConductor, nil
	case "arranger":
		return domain.RoleArranger, nil
	case "guest":
		return domain.RoleGuest, nil
	default:
		return domain.RoleGuest, fmt.Errorf("unknown role: %s", role)
	}
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

// SaveToJSON saves extracted album data to JSON format.
func SaveToJSON(albumData *AlbumData) ([]byte, error) {
	// Convert to domain album
	album, err := albumData.ToAlbum()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain: %w", err)
	}

	// Save using storage repository
	repo := storage.NewRepository()
	jsonData, err := repo.SaveToJSON(album)
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
