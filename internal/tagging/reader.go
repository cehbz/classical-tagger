package tagging

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/dhowden/tag"
)

// Metadata represents audio file metadata tags.
type Metadata struct {
	Title       string
	Artist      string
	Album       string
	Composer    string
	AlbumArtist string
	Year        string
	TrackNumber string
	DiscNumber  string
}

// Validate checks if required fields are present.
func (m Metadata) Validate() error {
	if m.Title == "" {
		return fmt.Errorf("title is required")
	}
	if m.Artist == "" {
		return fmt.Errorf("artist is required")
	}
	if m.Album == "" {
		return fmt.Errorf("album is required")
	}
	if m.TrackNumber == "" {
		return fmt.Errorf("track number is required")
	}
	// Composer is required for classical music
	if m.Composer == "" {
		return fmt.Errorf("composer is required for classical music")
	}
	return nil
}

// ToTrack converts metadata to a domain Track entity.
// The artist string is expected to be in the format: "Soloist, Ensemble, Conductor"
func (m Metadata) ToTrack(filename string) (*domain.Track, error) {
	// Parse track number
	trackNum, err := parseTrackNumber(m.TrackNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid track number %q: %w", m.TrackNumber, err)
	}

	// Parse disc number (default to 1)
	discNum := 1
	if m.DiscNumber != "" {
		discNum, err = strconv.Atoi(m.DiscNumber)
		if err != nil {
			return nil, fmt.Errorf("invalid disc number %q: %w", m.DiscNumber, err)
		}
	}

	// Create composer artist
	composer := domain.Artist{Name: m.Composer, Role: domain.RoleComposer}

	// Parse performers from Artist field
	// For now, we'll create a simple ensemble artist
	// TODO: Parse the "Soloist, Ensemble, Conductor" format properly
	artists := []domain.Artist{composer}
	if m.Artist != "" {
		// Simple implementation: treat entire Artist field as ensemble
		performer := domain.Artist{Name: m.Artist, Role: domain.RoleEnsemble}
		artists = append(artists, performer)
	}

	// Create track
	track := domain.Track{
		File: domain.File{
			Path: filename,
			Size: 0, // Size not available from reader
		},
		Disc:    discNum,
		Track:   trackNum,
		Title:   m.Title,
		Artists: artists,
	}

	return &track, nil
}

// parseTrackNumber handles various track number formats (e.g., "1", "01", "1/12")
func parseTrackNumber(s string) (int, error) {
	// Handle "track/total" format
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}

	return strconv.Atoi(strings.TrimSpace(s))
}

// FLACReader reads tags from FLAC files.
type FLACReader struct{}

// NewFLACReader creates a new FLACReader.
func NewFLACReader() *FLACReader {
	return &FLACReader{}
}

// ReadFile reads metadata from a FLAC file.
func (r *FLACReader) ReadFile(path string) (Metadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to read tags: %w", err)
	}

	track, _ := m.Track()
	disc, _ := m.Disc()

	metadata := Metadata{
		Title:       m.Title(),
		Artist:      m.Artist(),
		Album:       m.Album(),
		Composer:    m.Composer(),
		AlbumArtist: m.AlbumArtist(),
		Year:        strconv.Itoa(m.Year()),
		TrackNumber: strconv.Itoa(track),
		DiscNumber:  strconv.Itoa(disc),
	}

	return metadata, nil
}

// ReadTrackFromFile reads a FLAC file and returns a domain Track.
func (r *FLACReader) ReadTrackFromFile(path string, expectedDisc, expectedTrack int) (*domain.Track, error) {
	metadata, err := r.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Metadata: %w", err)
	}

	// Extract just the filename without path
	filename := path
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		filename = path[idx+1:]
	}
	if idx := strings.LastIndex(path, "\\"); idx != -1 {
		filename = path[idx+1:]
	}

	track, err := metadata.ToTrack(filename)
	if err != nil {
		return nil, err
	}

	// Verify disc and track numbers match expected
	if err := validateDiskAndTrackNumbers(track, expectedDisc, expectedTrack); err != nil {
		return nil, err
	}

	return track, nil
}

// validateDiskAndTrackNumbers checks that the track's disc/track match expectations.
// Separated for unit testing without needing real FLAC files.
func validateDiskAndTrackNumbers(track *domain.Track, expectedDisc, expectedTrack int) error {
	if track.Disc != expectedDisc {
		return fmt.Errorf("disc number mismatch: got %d, expected %d", track.Disc, expectedDisc)
	}
	if track.Track != expectedTrack {
		return fmt.Errorf("track number mismatch: got %d, expected %d", track.Track, expectedTrack)
	}
	return nil
}
