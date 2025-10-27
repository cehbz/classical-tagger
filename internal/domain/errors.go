package domain

import "errors"

// Standard domain errors
var (
	ErrEmptyArtistName = errors.New("artist name cannot be empty")
	ErrInvalidRole     = errors.New("invalid artist role")
)
