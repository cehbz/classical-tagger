// internal/uploader/domain.go
package uploader

import (
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// Torrent represents data from the Redacted torrent endpoint
type Torrent struct {
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

// TorrentGroup represents detailed data from the Redacted torrentgroup endpoint
type TorrentGroup struct {
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

// Metadata represents the final metadata ready for upload
type Metadata struct {
	// Core info
	Title string `json:"title"`
	Year  int    `json:"year"`

	Artists []domain.Artist `json:"artists"`

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

// Upload represents the final upload payload
type Upload struct {
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
	Importance []string `json:"importance[]"` // Must match artists[] length

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

// Note: "artists" role string appears with both importance 1 (main) and 2 (guest)
// So importance is 1:1 with role categories, but not with role strings.
var (
	// domainRoleToRedactedRole maps domain.Role to Redacted role string
	// Used for display/logging. For upload, use domainRoleToImportance instead.
	domainRoleToRedactedRole = map[domain.Role]string{
		domain.RoleComposer:  "composer",
		domain.RoleConductor: "conductor",
		domain.RoleEnsemble:  "artists",
		domain.RoleSoloist:   "artists",
		domain.RolePerformer: "artists",
		domain.RoleGuest:     "with",
		domain.RoleProducer:  "producer",
		domain.RoleDJ:        "dj",
		domain.RoleArranger:  "arranger",
		domain.RoleRemixer:   "remixer",
		domain.RoleUnknown:   "artists", // default to main artists
	}

	// domainRoleToImportance maps domain.Role to Redacted importance value
	// This is the primary mapping for uploads - all artists go in artists[] with importance[]
	domainRoleToImportance = map[domain.Role]string{
		domain.RoleComposer:  "4",
		domain.RoleConductor: "5",
		domain.RoleEnsemble:  "1",
		domain.RoleSoloist:   "1",
		domain.RolePerformer: "1",
		domain.RoleGuest:     "2",
		domain.RoleRemixer:   "3",
		domain.RoleProducer:  "7",
		domain.RoleDJ:        "6",
		domain.RoleArranger:  "8",
		domain.RoleUnknown:   "1",
	}

	// redactedRoleToDomainRole maps Redacted role strings to domain.Role
	// Handles variations like "composer"/"composers", "artists"/"artist"
	redactedRoleToDomainRole = map[string]domain.Role{
		"composer":   domain.RoleComposer,
		"composers":  domain.RoleComposer,
		"conductor":  domain.RoleConductor,
		"conductors": domain.RoleConductor,
		"artists":    domain.RolePerformer,
		"artist":     domain.RolePerformer,
		"with":       domain.RoleGuest,
		"guest":      domain.RoleGuest,
		"producer":   domain.RoleProducer,
		"dj":         domain.RoleDJ,
		"remixer":    domain.RoleRemixer,
		"remixedby":  domain.RoleRemixer,
		"arranger":   domain.RoleArranger,
		"arrangers":  domain.RoleArranger,
	}
)

// DomainRole converts Redacted artist role strings to our domain roles
func DomainRole(redactedRole string) domain.Role {
	if role, ok := redactedRoleToDomainRole[strings.ToLower(redactedRole)]; ok {
		return role
	}
	return domain.RoleUnknown // Default
}

// RedactedRole converts our domain roles to Redacted role strings
// Note: For uploads, use RedactedImportance() instead - all artists go in artists[]
// with importance[] values. This function is for display/logging purposes.
func RedactedRole(role domain.Role) string {
	if redactedRole, ok := domainRoleToRedactedRole[role]; ok {
		return redactedRole
	}
	return "artists" // Default
}

// RedactedImportance converts our domain roles to Redacted importance values
// This is the primary mapping for uploads - all artists go in artists[] with importance[]
func RedactedImportance(role domain.Role) string {
	if importance, ok := domainRoleToImportance[role]; ok {
		return importance
	}
	return "1" // Default to main artist
}
