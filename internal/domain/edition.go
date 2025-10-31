package domain

// Edition represents a specific release edition of an album.
// All fields are exported and mutable.
type Edition struct {
	Label         string `json:"label"`
	CatalogNumber string `json:"catalog_number,omitempty"`
	Year          int    `json:"year"`
}
