package scraping

// AlbumData represents extracted album metadata before domain conversion.
type AlbumData struct {
	Title        string        `json:"Title"`
	OriginalYear int           `json:"OriginalYear,omitempty"`
	Edition      *EditionData  `json:"Edition,omitempty"`
	Tracks       []TrackData   `json:"Tracks,omitempty"`
}

// EditionData represents edition information.
type EditionData struct {
	Label         string `json:"Label,omitempty"`
	CatalogNumber string `json:"CatalogNumber,omitempty"`
	EditionYear   int    `json:"EditionYear,omitempty"`
}

// TrackData represents a single track.
type TrackData struct {
	Disc     int          `json:"Disc,omitempty"`
	Track    int          `json:"Track,omitempty"`
	Title    string       `json:"Title"`
	Composer string       `json:"Composer,omitempty"`
	Artists  []ArtistData `json:"Artists,omitempty"`
	Arranger string       `json:"Arranger,omitempty"`
	Name     string       `json:"Name,omitempty"`
}

// ArtistData represents an artist.
type ArtistData struct {
	Name string `json:"Name"`
	Role string `json:"Role,omitempty"`
}

// Sentinel values for missing data
const (
	MissingTitle = "[Unknown Album]"
	MissingYear  = 0
)