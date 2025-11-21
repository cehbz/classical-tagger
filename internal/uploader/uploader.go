// internal/uploader/uploader.go
package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// UploadCommand handles the upload workflow
type UploadCommand struct {
	client       *RedactedClient
	cache        *Cache
	torrentDir   string
	torrentID    int
	trumpReason  string
	cacheDir     string
	dryRun       bool
	verbose      bool
}

// NewUploadCommand creates a new upload command
func NewUploadCommand(apiKey string, torrentDir string, torrentID int) *UploadCommand {
	// Determine cache directory using XDG standard
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache")
	}
	cacheDir = filepath.Join(cacheDir, "redacted-uploader")
	
	// Ensure cache directory exists
	os.MkdirAll(cacheDir, 0755)

	return &UploadCommand{
		client:     NewRedactedClient(apiKey),
		cache:      &Cache{dir: cacheDir, ttl: 24 * time.Hour},
		torrentDir: torrentDir,
		torrentID:  torrentID,
		cacheDir:   cacheDir,
	}
}

// SetTrumpReason sets a custom trump reason
func (c *UploadCommand) SetTrumpReason(reason string) {
	c.trumpReason = reason
}

// SetDryRun enables dry-run mode
func (c *UploadCommand) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// SetVerbose enables verbose output
func (c *UploadCommand) SetVerbose(verbose bool) {
	c.verbose = verbose
}

// Execute runs the upload workflow
func (c *UploadCommand) Execute(ctx context.Context) error {
	c.log("Starting upload workflow for torrent ID %d", c.torrentID)
	
	// Step 1: Fetch metadata from Redacted
	c.log("Fetching torrent metadata...")
	torrentMeta, err := c.fetchTorrentMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch torrent metadata: %w", err)
	}
	
	c.log("Fetching group metadata for group ID %d...", torrentMeta.GroupID)
	groupMeta, err := c.fetchGroupMetadata(ctx, torrentMeta.GroupID)
	if err != nil {
		return fmt.Errorf("failed to fetch group metadata: %w", err)
	}
	
	// Step 2: Load local metadata
	c.log("Loading local torrent metadata...")
	localTorrent, err := c.loadLocalTorrent()
	if err != nil {
		return fmt.Errorf("failed to load local torrent: %w", err)
	}
	
	// Step 3: Validate artists
	c.log("Validating artist consistency...")
	validationErrors := c.validateArtists(
		c.combineArtists(groupMeta),
		localTorrent.AlbumArtist,
	)
	
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "Validation error: %v\n", e)
		}
		if !c.dryRun {
			return fmt.Errorf("validation failed with %d errors", len(validationErrors))
		}
		c.log("Dry run mode - continuing despite validation errors")
	}
	
	// Step 4: Merge metadata
	c.log("Merging metadata...")
	trumpReason := c.trumpReason
	if trumpReason == "" {
		trumpReason = c.generateTrumpReason(localTorrent)
	}
	
	merged := c.mergeMetadata(torrentMeta, groupMeta, localTorrent, trumpReason)
	
	// Step 5: Validate required fields
	if err := c.validateRequiredFields(merged); err != nil {
		return fmt.Errorf("required field validation failed: %w", err)
	}
	
	// Step 6: Create torrent file
	c.log("Creating torrent file...")
	torrentPath, err := c.createTorrentFile(ctx, c.torrentDir, "https://flacsfor.me/announce")
	if err != nil {
		return fmt.Errorf("failed to create torrent file: %w", err)
	}
	
	// Step 7: Upload (or dry run)
	if c.dryRun {
		c.log("Dry run mode - would upload with the following metadata:")
		c.printMergedMetadata(merged)
		return nil
	}
	
	c.log("Uploading torrent...")
	uploadReq := c.prepareUploadRequest(merged)
	if err := c.client.UploadTorrent(ctx, uploadReq, torrentPath); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	
	c.log("Upload successful!")
	return nil
}

// fetchTorrentMetadata fetches torrent metadata with caching
func (c *UploadCommand) fetchTorrentMetadata(ctx context.Context) (*TorrentMetadata, error) {
	cacheKey := fmt.Sprintf("torrent_%d", c.torrentID)
	
	var cached CachedTorrentMetadata
	if ok, _ := c.cache.load(cacheKey, &cached); ok {
		c.log("Using cached torrent metadata")
		return &cached.Data, nil
	}
	
	meta, err := c.client.GetTorrent(ctx, c.torrentID)
	if err != nil {
		return nil, err
	}
	
	// Save to cache
	c.cache.save(cacheKey, &CachedTorrentMetadata{
		Timestamp: time.Now(),
		Data:      *meta,
	})
	
	return meta, nil
}

// fetchGroupMetadata fetches group metadata with caching
func (c *UploadCommand) fetchGroupMetadata(ctx context.Context, groupID int) (*GroupMetadata, error) {
	cacheKey := fmt.Sprintf("group_%d", groupID)
	
	var cached CachedGroupMetadata
	if ok, _ := c.cache.load(cacheKey, &cached); ok {
		c.log("Using cached group metadata")
		return &cached.Data, nil
	}
	
	meta, err := c.client.GetTorrentGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	
	// Save to cache
	c.cache.save(cacheKey, &CachedGroupMetadata{
		Timestamp: time.Now(),
		Data:      *meta,
	})
	
	return meta, nil
}

// loadLocalTorrent loads metadata from the local torrent directory
func (c *UploadCommand) loadLocalTorrent() (*domain.Torrent, error) {
	// Try to load from extracted JSON files
	torrent := &domain.Torrent{
		RootPath: c.torrentDir,
	}
	
	// Look for Discogs JSON
	discogsPath := filepath.Join(c.torrentDir, "discogs.json")
	if data, err := os.ReadFile(discogsPath); err == nil {
		if err := json.Unmarshal(data, torrent); err != nil {
			c.log("Warning: failed to parse discogs.json: %v", err)
		}
	}
	
	// Look for local extraction JSON
	localPath := filepath.Join(c.torrentDir, "metadata.json")
	if data, err := os.ReadFile(localPath); err == nil {
		var localMeta domain.Torrent
		if err := json.Unmarshal(data, &localMeta); err == nil {
			// Merge with priority to local
			if torrent.Title == "" {
				torrent.Title = localMeta.Title
			}
			if torrent.OriginalYear == 0 {
				torrent.OriginalYear = localMeta.OriginalYear
			}
			if len(torrent.Files) == 0 {
				torrent.Files = localMeta.Files
			}
		}
	}
	
	// If no metadata files, extract from FLAC files
	if torrent.Title == "" || len(torrent.Files) == 0 {
		if err := c.extractFromFLACs(torrent); err != nil {
			return nil, err
		}
	}
	
	return torrent, nil
}

// extractFromFLACs extracts metadata directly from FLAC files
func (c *UploadCommand) extractFromFLACs(torrent *domain.Torrent) error {
	// Walk directory to find FLAC files
	err := filepath.Walk(c.torrentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if strings.HasSuffix(strings.ToLower(path), ".flac") {
			relPath, _ := filepath.Rel(c.torrentDir, path)
			
			// Create a track entry
			track := &domain.Track{
				File: domain.File{
					Path: relPath,
				},
			}
			
			// TODO: Use go-flac to extract metadata
			// For now, we rely on pre-extracted JSON files
			
			torrent.Files = append(torrent.Files, track)
		}
		
		return nil
	})
	
	return err
}

// combineArtists combines all artist credits from group metadata
func (c *UploadCommand) combineArtists(group *GroupMetadata) []ArtistCredit {
	var artists []ArtistCredit
	
	for _, a := range group.Artists {
		a.Role = "artists"
		artists = append(artists, a)
	}
	for _, a := range group.Composers {
		a.Role = "composer"
		artists = append(artists, a)
	}
	for _, a := range group.Conductors {
		a.Role = "conductor"
		artists = append(artists, a)
	}
	for _, a := range group.With {
		a.Role = "with"
		artists = append(artists, a)
	}
	for _, a := range group.Producer {
		a.Role = "producer"
		artists = append(artists, a)
	}
	
	return artists
}

// validateArtists validates artist consistency between Redacted and local
func (c *UploadCommand) validateArtists(redacted []ArtistCredit, local []domain.Artist) []error {
	var errors []error
	
	// Build maps for comparison
	redactedMap := make(map[string]string) // name -> role
	for _, a := range redacted {
		redactedMap[a.Name] = a.Role
	}
	
	localMap := make(map[string]domain.Role) // name -> role
	for _, a := range local {
		localMap[a.Name] = a.Role
	}
	
	// Check each Redacted artist exists in local with matching role
	for _, ra := range redacted {
		localRole, exists := localMap[ra.Name]
		if !exists {
			errors = append(errors, fmt.Errorf("artist %q with role %q not found in local tags", ra.Name, ra.Role))
			continue
		}
		
		expectedRole := mapRedactedRoleToOurRole(ra.Role)
		if localRole != expectedRole {
			errors = append(errors, fmt.Errorf("artist %q role mismatch: Redacted has %q (mapped to %v), local has %v", 
				ra.Name, ra.Role, expectedRole, localRole))
		}
	}
	
	// Check for extra artists in local not in Redacted
	for name, role := range localMap {
		if _, exists := redactedMap[name]; !exists {
			errors = append(errors, fmt.Errorf("local artist %q with role %v not found in Redacted metadata", name, role))
		}
	}
	
	return errors
}

// mergeMetadata merges all metadata sources
func (c *UploadCommand) mergeMetadata(torrent *TorrentMetadata, group *GroupMetadata, local *domain.Torrent, trumpReason string) *MergedMetadata {
	merged := &MergedMetadata{
		// From local/extracted
		Title: local.Title,
		Year:  local.OriginalYear,
		
		// From Redacted group
		Artists:    group.Artists,
		Composers:  group.Composers,
		Conductors: group.Conductors,
		With:       group.With,
		Producer:   group.Producer,
		
		// From Redacted torrent
		Format:      torrent.Format,
		Encoding:    torrent.Encoding,
		Media:       torrent.Media,
		Tags:        torrent.Tags,
		GroupID:     torrent.GroupID,
		TorrentID:   torrent.TorrentID,
		
		// Remaster info
		Remastered:  torrent.Remastered,
		RemasterYear: torrent.RemasterYear,
		RemasterTitle: torrent.RemasterTitle,
		RemasterRecordLabel: torrent.RemasterRecordLabel,
		RemasterCatalogueNumber: torrent.RemasterCatalogueNumber,
		
		// Trump info
		TrumpReason: trumpReason,
	}
	
	// Add label/catalog from local if available
	if local.Edition != nil {
		merged.Label = local.Edition.Label
		merged.CatalogNumber = local.Edition.CatalogNumber
	}
	
	// Append trump reason to description
	merged.Description = torrent.Description
	if trumpReason != "" {
		merged.Description += "\n\n[Trump Upload] Fixed: " + trumpReason
	}
	
	return merged
}

// generateTrumpReason generates an automatic trump reason
func (c *UploadCommand) generateTrumpReason(torrent *domain.Torrent) string {
	// TODO: Analyze what was fixed based on validation results
	return "Corrected tags and filenames according to classical music guidelines"
}

// validateRequiredFields checks all required fields are present
func (c *UploadCommand) validateRequiredFields(meta *MergedMetadata) error {
	var missing []string
	
	if meta.Title == "" {
		missing = append(missing, "title")
	}
	if meta.Year == 0 {
		missing = append(missing, "year")
	}
	if meta.Format == "" {
		missing = append(missing, "format")
	}
	if meta.Encoding == "" {
		missing = append(missing, "encoding")
	}
	if meta.Media == "" {
		missing = append(missing, "media")
	}
	if len(meta.Tags) == 0 {
		missing = append(missing, "tags")
	}
	if len(meta.Artists) == 0 && len(meta.Composers) == 0 {
		missing = append(missing, "artists or composers")
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	
	return nil
}

// prepareUploadRequest converts merged metadata to upload request
func (c *UploadCommand) prepareUploadRequest(meta *MergedMetadata) *UploadRequest {
	req := &UploadRequest{
		Type:     "Music",
		GroupID:  meta.GroupID,
		Title:    meta.Title,
		Year:     meta.Year,
		Format:   meta.Format,
		Encoding: meta.Encoding,
		Media:    meta.Media,
		
		RecordLabel:     meta.Label,
		CatalogueNumber: meta.CatalogNumber,
		
		Remastered:      meta.Remastered,
		RemasterYear:    meta.RemasterYear,
		RemasterTitle:   meta.RemasterTitle,
		RemasterLabel:   meta.RemasterRecordLabel,
		RemasterCatalog: meta.RemasterCatalogueNumber,
		
		ReleaseDescription: meta.Description,
		Tags:              strings.Join(meta.Tags, ","),
		
		TrumpTorrent: meta.TorrentID,
		TrumpReason:  meta.TrumpReason,
	}
	
	// Convert artist credits to string arrays
	for _, a := range meta.Artists {
		req.Artists = append(req.Artists, a.Name)
	}
	for _, c := range meta.Composers {
		req.Composers = append(req.Composers, c.Name)
	}
	for _, c := range meta.Conductors {
		req.Conductors = append(req.Conductors, c.Name)
	}
	
	return req
}

// createTorrentFile creates a .torrent file
func (c *UploadCommand) createTorrentFile(ctx context.Context, sourceDir string, announceURL string) (string, error) {
	// Check cache first
	torrentPath := filepath.Join(c.cacheDir, fmt.Sprintf("torrent_%d.torrent", c.torrentID))
	if _, err := os.Stat(torrentPath); err == nil {
		c.log("Using cached torrent file")
		return torrentPath, nil
	}
	
	// Create torrent using mktorrent
	cmd := exec.CommandContext(ctx, "mktorrent",
		"-p",                    // Private torrent
		"-l", "18",             // Piece length 2^18 = 256KB
		"-a", announceURL,      // Announce URL
		"-o", torrentPath,      // Output file
		sourceDir,              // Source directory
	)
	
	if c.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mktorrent failed: %w", err)
	}
	
	return torrentPath, nil
}

// printMergedMetadata prints metadata for dry run
func (c *UploadCommand) printMergedMetadata(meta *MergedMetadata) {
	fmt.Printf("\n=== Upload Metadata ===\n")
	fmt.Printf("Title: %s\n", meta.Title)
	fmt.Printf("Year: %d\n", meta.Year)
	fmt.Printf("Format: %s / %s / %s\n", meta.Format, meta.Encoding, meta.Media)
	
	if meta.Label != "" || meta.CatalogNumber != "" {
		fmt.Printf("Label: %s - %s\n", meta.Label, meta.CatalogNumber)
	}
	
	if meta.Remastered {
		fmt.Printf("Remaster: %d - %s\n", meta.RemasterYear, meta.RemasterTitle)
	}
	
	fmt.Printf("\nArtists:\n")
	for _, a := range meta.Artists {
		fmt.Printf("  - %s\n", a.Name)
	}
	
	if len(meta.Composers) > 0 {
		fmt.Printf("\nComposers:\n")
		for _, c := range meta.Composers {
			fmt.Printf("  - %s\n", c.Name)
		}
	}
	
	if len(meta.Conductors) > 0 {
		fmt.Printf("\nConductors:\n")
		for _, c := range meta.Conductors {
			fmt.Printf("  - %s\n", c.Name)
		}
	}
	
	fmt.Printf("\nTags: %s\n", strings.Join(meta.Tags, ", "))
	fmt.Printf("\nTrump Reason: %s\n", meta.TrumpReason)
	fmt.Printf("\nDescription:\n%s\n", meta.Description)
}

// log logs a message if verbose mode is enabled
func (c *UploadCommand) log(format string, args ...interface{}) {
	if c.verbose {
		fmt.Printf("[UPLOAD] "+format+"\n", args...)
	}
}

// Cache implementation

// save saves data to cache
func (c *Cache) save(key string, data interface{}) error {
	path := filepath.Join(c.dir, key+".json")
	
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// load loads data from cache
func (c *Cache) load(key string, target interface{}) (bool, error) {
	path := filepath.Join(c.dir, key+".json")
	
	// Check if file exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	
	// Check if cache is expired
	if time.Since(info.ModTime()) > c.ttl {
		return false, nil
	}
	
	// Load file
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	
	if err := json.NewDecoder(file).Decode(target); err != nil {
		return false, err
	}
	
	return true, nil
}

// ClearCache removes all cached files
func (c *Cache) ClearCache() error {
	return filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".torrent") {
			return os.Remove(path)
		}
		return nil
	})
}