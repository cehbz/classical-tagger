package domain

import (
	"fmt"
	"strings"
)

// Edition is a value object representing a specific release edition of an album.
// It is immutable after creation.
type Edition struct {
	label         string
	catalogNumber string
	editionYear   int
}

// NewEdition creates a new Edition with the given label and edition year.
// Returns an error if the label is empty or the year is invalid.
func NewEdition(label string, editionYear int) (Edition, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return Edition{}, fmt.Errorf("edition label cannot be empty")
	}
	if editionYear <= 0 {
		return Edition{}, fmt.Errorf("edition year must be positive, got %d", editionYear)
	}
	
	return Edition{
		label:       label,
		editionYear: editionYear,
	}, nil
}

// WithCatalogNumber returns a new Edition with the catalog number set.
// This is a builder method that creates a copy.
func (e Edition) WithCatalogNumber(catalogNumber string) Edition {
	e.catalogNumber = strings.TrimSpace(catalogNumber)
	return e
}

// Label returns the record label.
func (e Edition) Label() string {
	return e.label
}

// CatalogNumber returns the catalog number.
func (e Edition) CatalogNumber() string {
	return e.catalogNumber
}

// Year returns the edition year.
func (e Edition) Year() int {
	return e.editionYear
}

// Validate checks the edition for completeness.
// Returns validation issues (warnings for missing optional fields).
func (e Edition) Validate() []ValidationIssue {
	var issues []ValidationIssue
	
	// Catalog number is optional but recommended
	if e.catalogNumber == "" {
		issues = append(issues, NewIssue(
			LevelWarning,
			0, // album-level
			"Classical Guide: Step 3",
			"Catalog number is recommended but missing",
		))
	}
	
	return issues
}
