package discogs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cehbz/classical-tagger/internal/cache"
	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/ratelimit"
)

// Client is a Discogs API client.
type Client struct {
	BaseURL     string
	Token       string
	HTTPClient  *http.Client
	RateLimiter *ratelimit.RateLimiter // Use shared rate limiter
	Cache       *cache.Cache           // Use shared cache
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

type Role string

// Artist represents an artist/performer.
type Artist struct {
	Name string `json:"name"`
	Role Role   `json:"role,omitempty"`
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
		BaseURL:     "https://api.discogs.com",
		Token:       token,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		RateLimiter: ratelimit.NewRateLimiter(60, time.Minute), // 60 per minute
		Cache:       cache.NewCache(0),
	}
}

// Search searches for releases by artist and album.
func (c *Client) Search(artist, album string) ([]*Release, error) {
	// Create a cache key from the query
	cacheKey := fmt.Sprintf("search_%s_%s", url.QueryEscape(artist), url.QueryEscape(album))

	// Try cache first
	var cached []*Release
	if c.Cache.LoadFrom(cacheKey, &cached, "discogs") {
		return cached, nil
	}

	// Rate limit
	ctx := context.Background()
	if err := c.RateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Build search URL
	u, err := url.Parse(c.BaseURL + "/database/search")
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
	req.Header.Set("Authorization", "Discogs token="+c.Token)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	c.RateLimiter.OnResponse()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discogs API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResp searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert search results to releases
	releases := make([]*Release, len(searchResp.Results))
	for i, result := range searchResp.Results {
		releases[i] = &Release{
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

	c.Cache.SaveTo(cacheKey, releases, "discogs")

	return releases, nil
}

// SearchSimple searches for releases using a simple query parameter.
// This is more forgiving than the advanced search with separate artist and release_title parameters.
// No format restriction is applied.
func (c *Client) SearchSimple(query string) ([]*Release, error) {
	// Create a cache key from the query
	cacheKey := fmt.Sprintf("search_simple_%s", url.QueryEscape(query))

	// Try cache first
	var cached []*Release
	if c.Cache.LoadFrom(cacheKey, &cached, "discogs") {
		return cached, nil
	}

	// Rate limit
	ctx := context.Background()
	if err := c.RateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Build search URL
	u, err := url.Parse(c.BaseURL + "/database/search")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("type", "release")
	// Note: No format restriction for fallback search
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add auth header
	req.Header.Set("Authorization", "Discogs token="+c.Token)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	c.RateLimiter.OnResponse()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discogs API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResp searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert search results to releases
	releases := make([]*Release, len(searchResp.Results))
	for i, result := range searchResp.Results {
		releases[i] = &Release{
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

	c.Cache.SaveTo(cacheKey, releases, "discogs")

	return releases, nil
}

// GetRelease fetches detailed information for a specific release.
func (c *Client) GetRelease(releaseID int) (*Release, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("release_%d", releaseID)
	var cached Release
	if c.Cache.LoadFrom(cacheKey, &cached, "discogs") {
		return &cached, nil
	}

	// Apply rate limiting
	ctx := context.Background()
	if err := c.RateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build URL
	u := fmt.Sprintf("%s/releases/%d", c.BaseURL, releaseID)

	// Create request
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	// Add auth header
	req.Header.Set("Authorization", "Discogs token="+c.Token)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	c.RateLimiter.OnResponse()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %d not found", releaseID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discogs API error: %d - %s", resp.StatusCode, string(body))
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

	c.Cache.SaveTo(cacheKey, release, "discogs")

	return &release, nil
}

type ArtistMap map[string]map[domain.Role]struct{}

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

// normalizeArtistName normalizes an artist name for comparison (case-insensitive)
func normalizeArtistName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// DomainRole determines the role for an artist with preference order:
// 1. Discogs main artist role (if present)
// 2. Discogs extraartists role (if artist name matches)
// 3. Local file metadata role (if artist name matches)
// 4. RoleUnknown (will cause error)
func (artist Artist) DomainRole(release *Release, localTorrent *domain.Torrent) domain.Role {
	// 1. Check if main artist has explicit role
	if role := artist.Role.DomainRole(); role != domain.RoleUnknown {
		return role
	}

	// 2. Check extraartists for matching name
	normalizedName := normalizeArtistName(artist.Name)
	if release != nil {
		for _, extraArtist := range release.ExtraArtists {
			if normalizeArtistName(extraArtist.Name) == normalizedName {
				if role := extraArtist.Role.DomainRole(); role != domain.RoleUnknown {
					return role
				}
			}
		}
	}

	// 3. Check local file metadata for matching name
	if localTorrent != nil {
		// Check album artists
		for _, localArtist := range localTorrent.AlbumArtist {
			if normalizeArtistName(localArtist.Name) == normalizedName {
				if localArtist.Role != domain.RoleUnknown {
					return localArtist.Role
				}
			}
		}
		// Check track artists
		for _, track := range localTorrent.Tracks() {
			for _, localArtist := range track.Artists {
				if normalizeArtistName(localArtist.Name) == normalizedName {
					if localArtist.Role != domain.RoleUnknown {
						return localArtist.Role
					}
				}
			}
		}
	}

	// 4. Infer from name
	if role := inferRoleFromName(artist.Name); role != domain.RoleUnknown {
		return role
	}

	// 5. Return unknown (will cause error)
	return domain.RoleUnknown
}

// DomainTorrent converts a Discogs Release to a domain Torrent
// localTorrent is optional and used to fill in missing role information from file metadata
func (release *Release) DomainTorrent(rootPath string, localTorrent *domain.Torrent) (*domain.Torrent, error) {

	if release == nil {
		return nil, fmt.Errorf("release is nil")
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

	// Add main artists (typically performers) with role determination
	for _, discogArtist := range release.Artists {
		role := discogArtist.DomainRole(release, localTorrent)
		albumArtistMap.Add(discogArtist.Name, role)
	}

	// Add extraartists with role determination
	for _, discogArtist := range release.ExtraArtists {
		role := discogArtist.DomainRole(release, localTorrent)
		albumArtistMap.Add(discogArtist.Name, role)
	}

	// Convert map to slice
	albumArtists := albumArtistMap.Artists()

	// Validate no unknown roles in album artists
	for _, artist := range albumArtists {
		if artist.Role == domain.RoleUnknown {
			return nil, fmt.Errorf("cannot determine role for album artist '%s'. Discogs has no role, extraartists has no matching entry, and file metadata has no matching entry", artist.Name)
		}
	}

	// Convert tracks
	tracks := make([]domain.FileLike, 0, len(release.Tracklist))
	for _, discogsTrack := range release.Tracklist {

		trackArtistsMap := albumArtistMap.Copy()

		// add all track artists to track with role determination
		for _, artist := range discogsTrack.Artists {
			role := artist.DomainRole(release, localTorrent)
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
				role := artist.DomainRole(release, localTorrent)
				subTrackArtistsMap.Add(artist.Name, role)
			}
			subTrackArtists := subTrackArtistsMap.Artists()

			// Validate no unknown roles in subtrack artists
			for _, artist := range subTrackArtists {
				if artist.Role == domain.RoleUnknown {
					return nil, fmt.Errorf("cannot determine role for track artist '%s' in subtrack '%s'. Discogs has no role, extraartists has no matching entry, and file metadata has no matching entry", artist.Name, subTrackTitle)
				}
			}
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

		trackArtists := trackArtistsMap.Artists()

		// Validate no unknown roles in track artists
		for _, artist := range trackArtists {
			if artist.Role == domain.RoleUnknown {
				return nil, fmt.Errorf("cannot determine role for track artist '%s' in track '%s'. Discogs has no role, extraartists has no matching entry, and file metadata has no matching entry", artist.Name, discogsTrack.Title)
			}
		}

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

	return torrent, nil
}

func (role Role) DomainRole() domain.Role {
	switch strings.ToLower(strings.TrimSpace(string(role))) {
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
	default:
		return domain.RoleUnknown
	}
}

// inferRoleFromName tries to determine the role of an artist from their name
// Returns domain.RoleUnknown if no role can be determined
func inferRoleFromName(name string) domain.Role {
	ensembleKeywords := map[string]Role{
		"orchestra":    "ensemble",
		"orchestre":    "ensemble",
		"orchester":    "ensemble",
		"philharmonic": "ensemble",
		"symphony":     "ensemble",
		"choir":        "ensemble",
		"chorus":       "ensemble",
		"kammerchor":   "ensemble",
		"ensemble":     "ensemble",
		"quartet":      "ensemble",
		"trio":         "ensemble",
		"quintet":      "ensemble",
		"sextet":       "ensemble",
		"consort":      "ensemble",
		"academy":      "ensemble",
		"chamber":      "ensemble",
	}
	for _, field := range strings.FieldsFunc(name, func(r rune) bool { return !unicode.IsLetter(r) }) {
		if role, ok := ensembleKeywords[strings.ToLower(field)]; ok {
			return role.DomainRole()
		}
	}
	return domain.RoleUnknown
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
func (release *Release) SaveToFile(filename string, rootPath string, localTorrent *domain.Torrent) error {
	// Convert to domain Torrent format
	torrent, err := release.DomainTorrent(rootPath, localTorrent)
	if err != nil {
		return fmt.Errorf("failed to convert Discogs release: %w", err)
	}
	if torrent == nil {
		return fmt.Errorf("failed to convert Discogs release: returned nil")
	}

	return torrent.Save(filename)
}
