// internal/uploader/client.go
package uploader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/cehbz/classical-tagger/internal/cache"
	"github.com/cehbz/classical-tagger/internal/ratelimit"
)

// RedactedClient handles API communication with Redacted
type RedactedClient struct {
	BaseURL     string
	APIKey      string
	HTTPClient  *http.Client
	RateLimiter *ratelimit.RateLimiter // Reuse the existing rate limiter
	Cache       *cache.Cache
}

// NewRedactedClient creates a new Redacted API client
func NewRedactedClient(apiKey string) *RedactedClient {
	return &RedactedClient{
		BaseURL:     "https://redacted.sh",
		APIKey:      apiKey,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		RateLimiter: ratelimit.NewRateLimiter(10, 10*time.Second), // 10 requests per 10 seconds
		Cache:       cache.NewCache(0),
	}
}

// GetTorrent fetches torrent metadata from Redacted
func (c *RedactedClient) GetTorrent(ctx context.Context, torrentID int) (*Torrent, error) {
	// Create a cache key from the torrent ID
	cacheKey := fmt.Sprintf("torrent_%d", torrentID)

	// Try cache first
	var cached Torrent
	if c.Cache.LoadFrom(cacheKey, &cached, "redacted") {
		return &cached, nil
	}
	// Apply rate limiting
	if err := c.RateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build URL
	u, err := url.Parse(c.BaseURL + "/ajax.php")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("action", "torrent")
	q.Set("id", strconv.Itoa(torrentID))
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add API key header
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	c.RateLimiter.OnResponse()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return nil, fmt.Errorf("rate limited, retry after %s seconds", retryAfter)
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp struct {
		Status   string `json:"status"`
		Error    string `json:"error,omitempty"`
		Response struct {
			Group struct {
				GroupID   int      `json:"groupId"`
				GroupName string   `json:"groupName"`
				GroupYear int      `json:"groupYear"`
				Tags      []string `json:"tags"`
			} `json:"group"`
			Torrent struct {
				ID                      int    `json:"id"`
				Format                  string `json:"format"`
				Encoding                string `json:"encoding"`
				Media                   string `json:"media"`
				Remastered              bool   `json:"remastered"`
				RemasterYear            int    `json:"remasterYear"`
				RemasterTitle           string `json:"remasterTitle"`
				RemasterRecordLabel     string `json:"remasterRecordLabel"`
				RemasterCatalogueNumber string `json:"remasterCatalogueNumber"`
				Description             string `json:"description"`
				FileList                string `json:"fileList"`
				Size                    int64  `json:"size"`
			} `json:"torrent"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	// Convert to our domain model
	metadata := &Torrent{
		GroupID:                 apiResp.Response.Group.GroupID,
		GroupName:               apiResp.Response.Group.GroupName,
		GroupYear:               apiResp.Response.Group.GroupYear,
		Tags:                    apiResp.Response.Group.Tags,
		TorrentID:               apiResp.Response.Torrent.ID,
		Format:                  apiResp.Response.Torrent.Format,
		Encoding:                apiResp.Response.Torrent.Encoding,
		Media:                   apiResp.Response.Torrent.Media,
		Remastered:              apiResp.Response.Torrent.Remastered,
		RemasterYear:            apiResp.Response.Torrent.RemasterYear,
		RemasterTitle:           apiResp.Response.Torrent.RemasterTitle,
		RemasterRecordLabel:     apiResp.Response.Torrent.RemasterRecordLabel,
		RemasterCatalogueNumber: apiResp.Response.Torrent.RemasterCatalogueNumber,
		Description:             apiResp.Response.Torrent.Description,
		FileList:                apiResp.Response.Torrent.FileList,
		Size:                    apiResp.Response.Torrent.Size,
	}

	c.Cache.SaveTo(cacheKey, metadata, "redacted")

	return metadata, nil
}

// GetTorrentGroup fetches detailed group metadata from Redacted
func (c *RedactedClient) GetTorrentGroup(ctx context.Context, groupID int) (*TorrentGroup, error) {
	// Create a cache key from the group ID
	cacheKey := fmt.Sprintf("group_%d", groupID)

	// Try cache first
	var cached TorrentGroup
	if c.Cache.LoadFrom(cacheKey, &cached, "redacted") {
		return &cached, nil
	}

	// Apply rate limiting
	if err := c.RateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build URL
	u, err := url.Parse(c.BaseURL + "/ajax.php")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("action", "torrentgroup")
	q.Set("id", strconv.Itoa(groupID))
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add API key header
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	c.RateLimiter.OnResponse()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return nil, fmt.Errorf("rate limited, retry after %s seconds", retryAfter)
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp struct {
		Status   string `json:"status"`
		Error    string `json:"error,omitempty"`
		Response struct {
			Group struct {
				ID            int      `json:"id"`
				Name          string   `json:"name"`
				Year          int      `json:"year"`
				Tags          []string `json:"tags"`
				WikiBody      string   `json:"wikiBody"`
				MusicBrainzID string   `json:"musicBrainzId"`
				VanityHouse   bool     `json:"vanityHouse"`
				MusicInfo     struct {
					Artists   []ArtistCredit `json:"artists"`
					Composers []ArtistCredit `json:"composers"`
					Conductor []ArtistCredit `json:"conductor"`
					With      []ArtistCredit `json:"with"`
					RemixedBy []ArtistCredit `json:"remixedBy"`
					Producer  []ArtistCredit `json:"producer"`
					DJ        []ArtistCredit `json:"dj"`
				} `json:"musicInfo"`
			} `json:"group"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	// Convert to our domain model
	metadata := &TorrentGroup{
		ID:            apiResp.Response.Group.ID,
		Name:          apiResp.Response.Group.Name,
		Year:          apiResp.Response.Group.Year,
		Artists:       apiResp.Response.Group.MusicInfo.Artists,
		Composers:     apiResp.Response.Group.MusicInfo.Composers,
		Conductors:    apiResp.Response.Group.MusicInfo.Conductor,
		With:          apiResp.Response.Group.MusicInfo.With,
		RemixedBy:     apiResp.Response.Group.MusicInfo.RemixedBy,
		Producer:      apiResp.Response.Group.MusicInfo.Producer,
		DJ:            apiResp.Response.Group.MusicInfo.DJ,
		Tags:          apiResp.Response.Group.Tags,
		WikiBody:      apiResp.Response.Group.WikiBody,
		MusicBrainzID: apiResp.Response.Group.MusicBrainzID,
		VanityHouse:   apiResp.Response.Group.VanityHouse,
	}

	c.Cache.SaveTo(cacheKey, metadata, "redacted")

	return metadata, nil
}

// Upload uploads a new torrent to Redacted
func (c *RedactedClient) Upload(ctx context.Context, upload *Upload, torrentFilePath string) error {
	// Do not cache upload requests

	// Apply rate limiting
	if err := c.RateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	// Read torrent file
	torrentData, err := os.ReadFile(torrentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read torrent file: %w", err)
	}

	// Create multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add torrent file
	fw, err := w.CreateFormFile("file_input", "upload.torrent")
	if err != nil {
		return err
	}
	if _, err := fw.Write(torrentData); err != nil {
		return err
	}

	// Add form fields
	fields := map[string]string{
		"type":         "Music",
		"groupid":      strconv.Itoa(upload.GroupID),
		"title":        upload.Title,
		"year":         strconv.Itoa(upload.Year),
		"format":       upload.Format,
		"bitrate":      upload.Encoding,
		"media":        upload.Media,
		"release_desc": upload.ReleaseDescription,
		"tags":         upload.Tags,
	}

	// Add optional fields
	if upload.RecordLabel != "" {
		fields["releasename"] = upload.RecordLabel
	}
	if upload.CatalogueNumber != "" {
		fields["cataloguenumber"] = upload.CatalogueNumber
	}

	// Add remaster fields if applicable
	if upload.Remastered {
		fields["remaster"] = "on"
		if upload.RemasterYear > 0 {
			fields["remaster_year"] = strconv.Itoa(upload.RemasterYear)
		}
		if upload.RemasterTitle != "" {
			fields["remaster_title"] = upload.RemasterTitle
		}
		if upload.RemasterLabel != "" {
			fields["remaster_record_label"] = upload.RemasterLabel
		}
		if upload.RemasterCatalog != "" {
			fields["remaster_catalogue_number"] = upload.RemasterCatalog
		}
	}

	// Add trump fields if applicable
	if upload.TrumpTorrent > 0 {
		fields["trump_torrent"] = strconv.Itoa(upload.TrumpTorrent)
		fields["trump_reason"] = upload.TrumpReason
	}

	// Write all fields
	for key, val := range fields {
		if err := w.WriteField(key, val); err != nil {
			return err
		}
	}

	// Add artists arrays
	for i, artist := range upload.Artists {
		if err := w.WriteField(fmt.Sprintf("artists[%d]", i), artist); err != nil {
			return err
		}
		if err := w.WriteField(fmt.Sprintf("importance[%d]", i), "1"); err != nil {
			return err
		}
	}

	// Add composers
	for i, composer := range upload.Composers {
		if err := w.WriteField(fmt.Sprintf("composers[%d]", i), composer); err != nil {
			return err
		}
	}

	// Add conductors
	for i, conductor := range upload.Conductors {
		if err := w.WriteField(fmt.Sprintf("conductors[%d]", i), conductor); err != nil {
			return err
		}
	}

	// Close multipart writer
	if err := w.Close(); err != nil {
		return err
	}

	// Create HTTP request
	uploadURL := c.BaseURL + "/upload.php"
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("User-Agent", "ClassicalTagger/1.0")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	c.RateLimiter.OnResponse()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}