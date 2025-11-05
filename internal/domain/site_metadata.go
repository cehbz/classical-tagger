package domain

// SiteMetadata represents site-specific metadata for a torrent.
// This is optional metadata that comes from the torrent site (e.g., Redacted, What.CD).
// All fields are exported and mutable.
type SiteMetadata struct {
	TorrentID   int      `json:"torrent_id"`
	GroupID     int      `json:"group_id"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	CoverArtURL string   `json:"cover_art_url,omitempty"`

	// Format describes the files
	Media    string `json:"media"`    // "CD", "WEB", "Vinyl"
	Format   string `json:"format"`   // "FLAC", "MP3"
	Encoding string `json:"encoding"` // "Lossless", "320", "V0"
	Scene    bool   `json:"scene"`
	HasLog   bool   `json:"has_log"`
	HasCue   bool   `json:"has_cue"`
	LogScore int    `json:"log_score"`

	ReleaseType int    `json:"release_type"`
	AnnounceURL string `json:"announce_url,omitempty"`
}
