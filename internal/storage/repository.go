package storage

import (
	"encoding/json"
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumDTO is the data transfer object for album JSON serialization.
type AlbumDTO struct {
	Title        string      `json:"title"`
	OriginalYear int         `json:"original_year"`
	Edition      *EditionDTO `json:"edition,omitempty"`
	Tracks       []TrackDTO  `json:"tracks"`
}

// EditionDTO represents edition information in JSON.
type EditionDTO struct {
	Label         string `json:"label"`
	CatalogNumber string `json:"catalog_number,omitempty"`
	EditionYear   int    `json:"edition_year"`
}

// TrackDTO represents a track in JSON.
type TrackDTO struct {
	Disc     int         `json:"disc"`
	Track    int         `json:"track"`
	Title    string      `json:"title"`
	Composer ArtistDTO   `json:"composer"`
	Artists  []ArtistDTO `json:"artists"`
	Name     string      `json:"name"`
}

// ArtistDTO represents an artist in JSON.
type ArtistDTO struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// ToAlbum converts an AlbumDTO to a domain Album.
func (dto AlbumDTO) ToAlbum() (*domain.Album, error) {
	// Create album
	album, err := domain.NewAlbum(dto.Title, dto.OriginalYear)
	if err != nil {
		return nil, fmt.Errorf("failed to create album: %w", err)
	}
	
	// Add edition if present
	if dto.Edition != nil {
		edition, err := domain.NewEdition(dto.Edition.Label, dto.Edition.EditionYear)
		if err != nil {
			return nil, fmt.Errorf("failed to create edition: %w", err)
		}
		if dto.Edition.CatalogNumber != "" {
			edition = edition.WithCatalogNumber(dto.Edition.CatalogNumber)
		}
		album = album.WithEdition(edition)
	}
	
	// Add tracks
	for i, trackDTO := range dto.Tracks {
		// Parse composer
		composerRole, err := domain.ParseRole(trackDTO.Composer.Role)
		if err != nil {
			return nil, fmt.Errorf("track %d: invalid composer role: %w", i+1, err)
		}
		composer, err := domain.NewArtist(trackDTO.Composer.Name, composerRole)
		if err != nil {
			return nil, fmt.Errorf("track %d: invalid composer: %w", i+1, err)
		}
		
		// Parse other artists
		artists := []domain.Artist{composer}
		for _, artistDTO := range trackDTO.Artists {
			role, err := domain.ParseRole(artistDTO.Role)
			if err != nil {
				return nil, fmt.Errorf("track %d: invalid artist role %q: %w", i+1, artistDTO.Role, err)
			}
			artist, err := domain.NewArtist(artistDTO.Name, role)
			if err != nil {
				return nil, fmt.Errorf("track %d: invalid artist: %w", i+1, err)
			}
			artists = append(artists, artist)
		}
		
		// Create track
		track, err := domain.NewTrack(trackDTO.Disc, trackDTO.Track, trackDTO.Title, artists)
		if err != nil {
			return nil, fmt.Errorf("track %d: failed to create track: %w", i+1, err)
		}
		
		if trackDTO.Name != "" {
			track = track.WithName(trackDTO.Name)
		}
		
		if err := album.AddTrack(track); err != nil {
			return nil, fmt.Errorf("track %d: failed to add track: %w", i+1, err)
		}
	}
	
	return album, nil
}

// FromAlbum converts a domain Album to an AlbumDTO.
func FromAlbum(album *domain.Album) AlbumDTO {
	dto := AlbumDTO{
		Title:        album.Title(),
		OriginalYear: album.OriginalYear(),
		Tracks:       make([]TrackDTO, 0, len(album.Tracks())),
	}
	
	// Add edition if present
	if album.Edition() != nil {
		edition := album.Edition()
		dto.Edition = &EditionDTO{
			Label:         edition.Label(),
			CatalogNumber: edition.CatalogNumber(),
			EditionYear:   edition.Year(),
		}
	}
	
	// Add tracks
	for _, track := range album.Tracks() {
		composer := track.Composer()
		
		trackDTO := TrackDTO{
			Disc:  track.Disc(),
			Track: track.Track(),
			Title: track.Title(),
			Composer: ArtistDTO{
				Name: composer.Name(),
				Role: composer.Role().String(),
			},
			Artists: make([]ArtistDTO, 0),
			Name:    track.Name(),
		}
		
		// Add non-composer artists
		for _, artist := range track.Artists() {
			if artist.Role() != domain.RoleComposer {
				trackDTO.Artists = append(trackDTO.Artists, ArtistDTO{
					Name: artist.Name(),
					Role: artist.Role().String(),
				})
			}
		}
		
		dto.Tracks = append(dto.Tracks, trackDTO)
	}
	
	return dto
}

// Repository handles JSON serialization and deserialization of albums.
type Repository struct{}

// NewRepository creates a new Repository.
func NewRepository() *Repository {
	return &Repository{}
}

// SaveToJSON serializes an album to JSON bytes.
func (r *Repository) SaveToJSON(album *domain.Album) ([]byte, error) {
	dto := FromAlbum(album)
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal album: %w", err)
	}
	return data, nil
}

// LoadFromJSON deserializes an album from JSON bytes.
func (r *Repository) LoadFromJSON(data []byte) (*domain.Album, error) {
	var dto AlbumDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	album, err := dto.ToAlbum()
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTO to album: %w", err)
	}
	
	return album, nil
}

// SaveToFile saves an album to a JSON file.
func (r *Repository) SaveToFile(album *domain.Album, path string) error {
	data, err := r.SaveToJSON(album)
	if err != nil {
		return err
	}
	
	// TODO: Write to file
	// For now, just validate that we can serialize
	_ = data
	return fmt.Errorf("SaveToFile not yet implemented")
}

// LoadFromFile loads an album from a JSON file.
func (r *Repository) LoadFromFile(path string) (*domain.Album, error) {
	// TODO: Read from file
	return nil, fmt.Errorf("LoadFromFile not yet implemented")
}
