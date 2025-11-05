package scraping

import "github.com/cehbz/classical-tagger/internal/domain"

// ExtractionError represents a single error that occurred during extraction.
type ExtractionError struct {
	Field    string
	Message  string
	Required bool
}

// ExtractionResult wraps the extracted torrent data with metadata about the extraction.
type ExtractionResult struct {
	Torrent    *domain.Torrent   // Torrent is the extracted torrent data
	Source     string            // Source is the URL or identifier of the source
	Confidence float64           // Confidence indicates how confident we are in the extraction (0.0-1.0)
	Errors     []ExtractionError // Errors contains any errors encountered during extraction
	Warnings   []string          // Warnings contains any issues encountered during extraction
	Notes      []string          // Notes contains additional information about the extraction
}
