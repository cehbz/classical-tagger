package validation

import "github.com/cehbz/classical-tagger/internal/domain"

// AlbumValidator validates album metadata against rules.
type AlbumValidator struct{}

// NewAlbumValidator creates a new AlbumValidator.
func NewAlbumValidator() *AlbumValidator {
	return &AlbumValidator{}
}

// ValidateMetadata validates an album's metadata.
// This delegates to the domain's Validate() method which already
// implements all the metadata validation rules.
func (v *AlbumValidator) ValidateMetadata(album *domain.Album) []domain.ValidationIssue {
	return album.Validate()
}
