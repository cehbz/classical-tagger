package discogs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

const defaultBaseURL = "https://api.discogs.com"

// Client is a Discogs API client.
type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// Release represents a Discogs release.
type Release struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	Year          int      `json:"year"`
	Label         string   `json:"label,omitempty"`
	CatalogNumber string   `json:"catalog_number,omitempty"`
	Country       string   `json:"country,omitempty"`
	Format        []string `json:"format,omitempty"`
	Artists       []Artist `json:"artists,omitempty"`
	ExtraArtists  []Artist `json:"extraartists,omitempty"`
	Tracklist     []Track  `json:"tracklist,omitempty"`
	Labels        []Label  `json:"labels,omitempty"`
}

// Artist represents an artist/performer.
type Artist struct {
	Name string `json:"name"`
	Role string `json:"role,omitempty"`
}

// Track represents a track in the release.
type Track struct {
	Position  string   `json:"position"`
	Title     string   `json:"title"`
	Duration  string   `json:"duration,omitempty"`
	Artists   []Artist `json:"extraartists,omitempty"`
	SubTracks []Track  `json:"sub_tracks,omitempty"` // Subtracks for hierarchical works
}

// Label represents label information.
type Label struct {
	Name          string `json:"name"`
	CatalogNumber string `json:"catno"`
}

// searchResponse represents the Discogs search API response.
type searchResponse struct {
	Results []searchResult `json:"results"`
}

// searchResult represents a single search result.
type searchResult struct {
	ID      int      `json:"id"`
	Title   string   `json:"title"`
	Year    string   `json:"year,omitempty"`
	Label   []string `json:"label,omitempty"`
	Catno   string   `json:"catno,omitempty"`
	Format  []string `json:"format,omitempty"`
	Country string   `json:"country,omitempty"`
}

// NewClient creates a new Discogs API client.
func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: defaultBaseURL,
		http:    &http.Client{},
	}
}

// Search searches for releases by artist and album.
func (c *Client) Search(artist, album string) ([]Release, error) {
	// Build search URL
	u, err := url.Parse(c.baseURL + "/database/search")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("artist", artist)
	q.Set("release_title", album)
	q.Set("type", "release")
	q.Set("format", "CD") // Prefer CD releases for classical music
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add auth header
	req.Header.Set("Authorization", "Discogs token="+c.token)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Discogs API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResp searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert search results to releases
	releases := make([]Release, len(searchResp.Results))
	for i, result := range searchResp.Results {
		releases[i] = Release{
			ID:            result.ID,
			Title:         result.Title,
			Country:       result.Country,
			CatalogNumber: result.Catno,
			Format:        result.Format,
		}

		// Parse year
		if result.Year != "" {
			if year, err := strconv.Atoi(result.Year); err == nil {
				releases[i].Year = year
			}
		}

		// Get first label if available
		if len(result.Label) > 0 {
			releases[i].Label = result.Label[0]
		}
	}

	return releases, nil
}

// GetRelease fetches detailed information for a specific release.
func (c *Client) GetRelease(releaseID int) (*Release, error) {
	// Build URL
	u := fmt.Sprintf("%s/releases/%d", c.baseURL, releaseID)

	// Create request
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	// Add auth header
	req.Header.Set("Authorization", "Discogs token="+c.token)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %d not found", releaseID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Discogs API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release response: %w", err)
	}

	// Extract label and catalog from labels array if present
	if len(release.Labels) > 0 {
		release.Label = release.Labels[0].Name
		release.CatalogNumber = release.Labels[0].CatalogNumber
	}

	return &release, nil
}

type ArtistMap map[string]map[domain.Role]struct{}

func (a *ArtistMap) removeUnknownRoles() {
	// get rid of any unknown roles if there's a known role for that name
	for _, roles := range *a {
		if len(roles) > 1 {
			delete(roles, domain.RoleUnknown)
		}
	}
}

func (a ArtistMap) Artists() []domain.Artist {
	artists := make([]domain.Artist, 0, len(a))
	for name, roles := range a {
		for role := range roles {
			artists = append(artists, domain.Artist{Name: name, Role: role})
		}
	}
	return artists
}

func (a ArtistMap) Copy() ArtistMap {
	newMap := make(ArtistMap)
	for name, roles := range a {
		newMap[name] = make(map[domain.Role]struct{})
		for role := range roles {
			newMap[name][role] = struct{}{}
		}
	}
	return newMap
}

func (a *ArtistMap) Add(name string, role domain.Role) {
	if (*a)[name] == nil {
		(*a)[name] = make(map[domain.Role]struct{})
	}
	(*a)[name][role] = struct{}{}
}

// convertDiscogsRelease converts a Discogs Release to a domain Torrent
func (release *Release) DomainTorrent(rootPath string) *domain.Torrent {

	if release == nil {
		return nil
	}

	// Convert edition
	var edition *domain.Edition
	if release.Label != "" || release.CatalogNumber != "" || release.Year > 0 {
		edition = &domain.Edition{
			Label:         release.Label,
			CatalogNumber: release.CatalogNumber,
			Year:          release.Year,
		}
	}

	// Convert album artists from main artists and extraartists
	// Use a map to deduplicate by name and role
	albumArtistMap := make(ArtistMap)

	// Add main artists (typically performers)
	for _, discogArtist := range release.Artists {
		albumArtistMap.Add(discogArtist.Name, discogArtist.DomainRole())
	}

	for _, discogArtist := range release.ExtraArtists {
		albumArtistMap.Add(discogArtist.Name, discogArtist.DomainRole())
	}

	// Convert map to slice
	albumArtistMap.removeUnknownRoles()
	albumArtists := albumArtistMap.Artists()

	// Convert tracks
	tracks := make([]domain.FileLike, 0, len(release.Tracklist))
	for _, discogsTrack := range release.Tracklist {

		trackArtistsMap := albumArtistMap.Copy()

		// add all track artists to track
		for _, artist := range discogsTrack.Artists {
			role := artist.DomainRole()
			trackArtistsMap.Add(artist.Name, role)
		}

		// Process any subtracks - these have explicit positions and titles
		for _, subtrack := range discogsTrack.SubTracks {
			subTrackDisc, subTrackNum := parseDiscogsPosition(subtrack.Position)
			if subTrackNum == 0 {
				// Invalid position, skip
				continue
			}

			// Build track title: prepend parent work title to subtrack title
			subTrackTitle := discogsTrack.Title + ": " + subtrack.Title

			// Build track artists: add parent composer
			subTrackArtistsMap := trackArtistsMap.Copy()

			for _, artist := range subtrack.Artists {
				role := artist.DomainRole()
				subTrackArtistsMap.Add(artist.Name, role)
			}
			subTrackArtistsMap.removeUnknownRoles()
			subTrackArtists := subTrackArtistsMap.Artists()
			// Generate a path from track number and title
			path := generateTrackPath(subTrackNum, subTrackTitle)

			domainSubTrack := &domain.Track{
				File: domain.File{
					Path: path,
				},
				Disc:    subTrackDisc,
				Track:   subTrackNum,
				Title:   subTrackTitle,
				Artists: subTrackArtists,
			}
			tracks = append(tracks, domainSubTrack)
		}

		disc, trackNum := parseDiscogsPosition(discogsTrack.Position)
		if trackNum == 0 {
			// Invalid position, skip
			continue
		}

		// Generate a path from track number and title (since we don't have actual files)
		path := generateTrackPath(trackNum, discogsTrack.Title)

		trackArtistsMap.removeUnknownRoles()
		trackArtists := trackArtistsMap.Artists()

		track := &domain.Track{
			File: domain.File{
				Path: path,
			},
			Disc:    disc,
			Track:   trackNum,
			Title:   discogsTrack.Title,
			Artists: trackArtists,
		}
		tracks = append(tracks, track)
	}

	torrent := &domain.Torrent{
		Title:        release.Title,
		OriginalYear: release.Year,
		Edition:      edition,
		AlbumArtist:  albumArtists,
		Files:        tracks,
		SiteMetadata: nil,
	}

	// Generate root_path using the same logic as directory naming
	torrent.RootPath = path.Join(rootPath, torrent.DirectoryName())

	return torrent
}

// mapDiscogsRoleToDomain maps Discogs role strings to domain Role enum
func (artist *Artist) DomainRole() domain.Role {
	roleLower := strings.ToLower(strings.TrimSpace(artist.Role))
	nameLower := strings.ToLower(artist.Name)

	// Map explicit Discogs roles
	switch roleLower {
	case "composed by", "composer":
		return domain.RoleComposer
	case "conductor", "conducted by", "chorus master":
		return domain.RoleConductor
	case "choir", "chorus", "orchestra", "orchestre", "orchester", "ensemble":
		return domain.RoleEnsemble
	case "soloist", "solo":
		return domain.RoleSoloist
	case "arranger", "arranged by":
		return domain.RoleArranger
	case "guest":
		return domain.RoleGuest
	}

	// Infer from name if role is empty or unknown
	if roleLower == "" {
		// Check for ensemble keywords in name
		ensembleKeywords := []string{
			"orchestra", "orchestre", "orchester", "philharmonic", "symphony",
			"choir", "chorus", "kammerchor", "ensemble", "quartet", "trio",
			"quintet", "sextet", "consort", "academy", "chamber",
		}
		for _, keyword := range ensembleKeywords {
			if strings.Contains(nameLower, keyword) {
				return domain.RoleEnsemble
			}
		}
	}

	return domain.RoleUnknown
}

func (artist *Artist) DomainArtist() domain.Artist {
	return domain.Artist{
		Name: artist.Name,
		Role: artist.DomainRole(),
	}
}

// parseDiscogsPosition parses a Discogs position string (e.g., "1", "1-1", "A1", "CD1-1")
// Returns (disc, track)
// TODO: Handle track as string
func parseDiscogsPosition(position string) (int, int) {
	position = strings.TrimSpace(position)
	if position == "" {
		return 1, 0
	}

	// Handle formats like "1", "1-1", "A1", "CD1-1", etc.
	parts := strings.Split(position, "-")
	if len(parts) == 1 {
		track, _ := strconv.Atoi(position)
		return 1, track
	}

	// Format: "disc-track" or "CD1-1"
	discStr := strings.TrimSpace(parts[0])
	trackStr := strings.TrimSpace(parts[1])

	// Remove non-numeric prefix from disc (e.g., "CD1" -> "1")
	discStr = regexp.MustCompile(`[^0-9]`).ReplaceAllString(discStr, "")
	disc, _ := strconv.Atoi(discStr)
	if disc == 0 {
		return 1, 0
	}
	trackStr = regexp.MustCompile(`[^0-9]`).ReplaceAllString(trackStr, "")
	track, _ := strconv.Atoi(trackStr)
	if track == 0 {
		return 1, 0
	}

	return disc, track
}

// generateTrackPath generates a file path from track number and title
func generateTrackPath(track int, title string) string {
	if track == 0 {
		return ""
	}

	// Sanitize title for filename
	sanitized := strings.ReplaceAll(title, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	sanitized = strings.TrimSpace(sanitized)

	return fmt.Sprintf("%02d-%s.flac", track, sanitized)
}

// saveDiscogsRelease converts Discogs release to domain Torrent and saves to file
func (release *Release) SaveToFile(filename string, rootPath string) error {
	// Convert to domain Torrent format
	torrent := release.DomainTorrent(rootPath)
	if torrent == nil {
		return fmt.Errorf("failed to convert Discogs release")
	}

	return torrent.Save(filename)
}
