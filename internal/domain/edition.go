package domain

// Edition represents a specific release edition of an album.
// All fields are exported and mutable.
type Edition struct {
	Label         string `json:"label"`
	CatalogNumber string `json:"catalog_number,omitempty"`
	Year          int    `json:"year"`
}

// Validate checks the edition for completeness.
// Returns validation issues (warnings for missing optional fields).
func (e *Edition) Validate() []ValidationIssue {
	var issues []ValidationIssue

	// Catalog number is optional but recommended
	if e.CatalogNumber == "" {
		issues = append(issues, ValidationIssue{
			Level:   LevelWarning,
			Track:   0, // album-level
			Rule:    "Classical Guide: Step 3",
			Message: "Catalog number is recommended but missing",
		})
	}

	return issues
}
