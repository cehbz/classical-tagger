// internal/uploader/domain.go
package uploader

import (
	"time"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TorrentMetadata represents data from the Redacted torrent endpoint
type TorrentMetadata struct {
	// Group info from torrent response
	GroupID   int      `json:"groupId"`
	GroupName string   `json:"groupName"`
	GroupYear int      `json:"groupYear"`
	Tags      []string `json:"tags"`

	// Torrent specific info
	TorrentID               int    `json:"torrentId"`
	Format                  string `json:"format"`
	Encoding                string `json:"encoding"`
	Media                   string `json:"media"`
	Remastered              bool   `json:"remastered"`
	RemasterYear            int    `json:"remasterYear,omitempty"`
	RemasterTitle           string `json:"remasterTitle,omitempty"`
	RemasterRecordLabel     string `json:"remasterRecordLabel,omitempty"`
	RemasterCatalogueNumber string `json:"remasterCatalogueNumber,omitempty"`
	Description             string `json:"description"`
	FileList                string `json:"fileList"`
	Size                    int64  `json:"size"`
}

// GroupMetadata represents detailed data from the Redacted torrentgroup endpoint
type GroupMetadata struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	Year          int            `json:"year"`
	Artists       []ArtistCredit `json:"artists"`
	Composers     []ArtistCredit `json:"composers"`
	Conductors    []ArtistCredit `json:"conductors"`
	With          []ArtistCredit `json:"with"` // Featured/guest artists
	RemixedBy     []ArtistCredit `json:"remixedBy"`
	Producer      []ArtistCredit `json:"producer"`
	DJ            []ArtistCredit `json:"dj"`
	Tags          []string       `json:"tags"`
	WikiBody      string         `json:"wikiBody"`
	MusicBrainzID string         `json:"musicBrainzId,omitempty"`
	VanityHouse   bool           `json:"vanityHouse"`
}

// ArtistCredit represents an artist with their role
type ArtistCredit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // "artists", "composer", "conductor", etc.
}

// MergedMetadata represents the final metadata ready for upload
type MergedMetadata struct {
	// Core info
	Title string `json:"title"`
	Year  int    `json:"year"`

	// Artists - from Redacted
	Artists    []ArtistCredit `json:"artists"`
	Composers  []ArtistCredit `json:"composers,omitempty"`
	Conductors []ArtistCredit `json:"conductors,omitempty"`
	With       []ArtistCredit `json:"with,omitempty"`
	Producer   []ArtistCredit `json:"producer,omitempty"`

	// Release info - from local files/Discogs
	Label         string `json:"recordLabel,omitempty"`
	CatalogNumber string `json:"catalogueNumber,omitempty"`

	// Format info - from Redacted
	Format                  string `json:"format"`
	Encoding                string `json:"encoding"`
	Media                   string `json:"media"`
	Remastered              bool   `json:"remastered"`
	RemasterYear            int    `json:"remasterYear,omitempty"`
	RemasterTitle           string `json:"remasterTitle,omitempty"`
	RemasterRecordLabel     string `json:"remasterRecordLabel,omitempty"`
	RemasterCatalogueNumber string `json:"remasterCatalogueNumber,omitempty"`

	// Site metadata - from Redacted
	Tags        []string `json:"tags"`
	Description string   `json:"description"`

	// Upload specific
	TrumpReason string `json:"trumpReason"`
	GroupID     int    `json:"groupId"`
	TorrentID   int    `json:"torrentId"` // ID being trumped
}

// UploadRequest represents the final upload payload
type UploadRequest struct {
	// Torrent file
	TorrentFile []byte `json:"-"` // Binary data, not JSON

	// Metadata fields for the upload form
	Type     string `json:"type"`    // "Music"
	GroupID  int    `json:"groupid"` // Existing group to add to
	Title    string `json:"title"`   // Album title
	Year     int    `json:"year"`    // Original year
	Format   string `json:"format"`  // "FLAC"
	Encoding string `json:"bitrate"` // "Lossless", "24bit Lossless", etc.
	Media    string `json:"media"`   // "CD", "WEB", "Vinyl", etc.

	// Release info
	RecordLabel     string `json:"releasename,omitempty"` // Label name
	CatalogueNumber string `json:"cataloguenumber,omitempty"`

	// Remaster info
	Remastered      bool   `json:"remaster"`
	RemasterYear    int    `json:"remaster_year,omitempty"`
	RemasterTitle   string `json:"remaster_title,omitempty"`
	RemasterLabel   string `json:"remaster_record_label,omitempty"`
	RemasterCatalog string `json:"remaster_catalogue_number,omitempty"`

	// Artists - these are arrays of strings in the API
	Artists    []string `json:"artists[]"`
	Importance []string `json:"importance[]"` // "1" for main, "2" for guest, etc.
	Composers  []string `json:"composers[]"`
	Conductors []string `json:"conductors[]"`
	DJ         []string `json:"dj[]"`

	// Description and tags
	ReleaseDescription string `json:"release_desc"`
	Tags               string `json:"tags"` // Comma-separated

	// Trump specific
	VanityHouse  bool   `json:"vanity_house"`
	TrumpTorrent int    `json:"trump_torrent,omitempty"` // ID to trump
	TrumpReason  string `json:"trump_reason,omitempty"`
}

// ValidationError represents an error during validation
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Cache represents the metadata cache
type Cache struct {
	dir string
	ttl time.Duration
}

// CachedTorrentMetadata wraps metadata with timestamp
type CachedTorrentMetadata struct {
	Timestamp time.Time       `json:"timestamp"`
	Data      TorrentMetadata `json:"data"`
}

// CachedGroupMetadata wraps group metadata with timestamp
type CachedGroupMetadata struct {
	Timestamp time.Time     `json:"timestamp"`
	Data      GroupMetadata `json:"data"`
}

// mapRedactedRoleToOurRole converts Redacted artist roles to our domain roles
func mapRedactedRoleToOurRole(redactedRole string) domain.Role {
	switch redactedRole {
	case "composer", "composers":
		return domain.RoleComposer
	case "conductor", "conductors":
		return domain.RoleConductor
	case "artists", "artist":
		return domain.RoleEnsemble // Or RolePerformer based on context
	case "with", "guest":
		return domain.RolePerformer
	case "producer":
		return domain.RoleProducer
	case "dj":
		return domain.RolePerformer
	default:
		return domain.RoleUnknown
	}
}

// mapOurRoleToRedactedRole converts our domain roles to Redacted roles
func mapOurRoleToRedactedRole(role domain.Role) string {
	switch role {
	case domain.RoleComposer:
		return "composer"
	case domain.RoleConductor:
		return "conductor"
	case domain.RoleEnsemble:
		return "artists"
	case domain.RolePerformer, domain.RoleSoloist:
		return "artists"
	case domain.RoleProducer:
		return "producer"
	default:
		return "artists"
	}
}
