// internal/uploader/uploader.go
package uploader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cehbz/classical-tagger/internal/cache"
	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/tagging"
)

// UploadCommand handles the upload workflow
type UploadCommand struct {
	Client      *RedactedClient
	Cache       *cache.Cache // Reuse common cache implementation
	TorrentDir  string
	TorrentID   int
	TrumpReason string
	CacheDir    string
	DryRun      bool
	Verbose     bool
}

// NewUploadCommand creates a new upload command
func NewUploadCommand(apiKey string, torrentDir string, torrentID int) *UploadCommand {
	// Use common cache implementation
	cacheImpl := cache.NewCache(0)

	return &UploadCommand{
		Client:     NewRedactedClient(apiKey),
		Cache:      cacheImpl,
		TorrentDir: torrentDir,
		TorrentID:  torrentID,
		CacheDir:   cacheImpl.GetCacheDir("redacted-uploader"),
	}
}

// Execute runs the upload workflow
func (c *UploadCommand) Execute(ctx context.Context) error {
	c.log("Starting upload workflow for torrent ID %d", c.TorrentID)

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

	// Step 3: Validate that local artists are a superset of Redacted artists
	c.log("Validating artist consistency...")
	allLocalArtists := c.collectAllLocalArtists(localTorrent)
	redactedArtists := c.combineArtists(groupMeta)
	validationErrors := c.validateArtistsSuperset(redactedArtists, allLocalArtists)

	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "Validation error: %v\n", e)
		}
		if !c.DryRun {
			return fmt.Errorf("validation failed with %d errors", len(validationErrors))
		}
		c.log("Dry run mode - continuing despite validation errors")
	}

	// Step 4: Merge metadata
	c.log("Merging metadata...")
	trumpReason := c.TrumpReason
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
	torrentPath, err := c.createTorrentFile(ctx, c.TorrentDir, "https://flacsfor.me/announce")
	if err != nil {
		return fmt.Errorf("failed to create torrent file: %w", err)
	}

	// Step 7: Upload (or dry run)
	if c.DryRun {
		c.log("Dry run mode - would upload with the following metadata:")
		c.printMergedMetadata(merged)
		return nil
	}

	c.log("Uploading torrent...")
	uploadReq := c.prepareUploadRequest(merged)
	if err := c.Client.Upload(ctx, uploadReq, torrentPath); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	c.log("Upload successful!")
	return nil
}

// fetchTorrentMetadata fetches torrent metadata with caching
func (c *UploadCommand) fetchTorrentMetadata(ctx context.Context) (*Torrent, error) {
	cacheKey := fmt.Sprintf("torrent_%d", c.TorrentID)

	var cached Torrent
	if c.Cache.LoadFrom(cacheKey, &cached, "redacted") {
		c.log("Using cached torrent metadata")
		return &cached, nil
	}

	meta, err := c.Client.GetTorrent(ctx, c.TorrentID)
	if err != nil {
		return nil, err
	}

	// Save to cache
	c.Cache.SaveTo(cacheKey, meta, "redacted")

	return meta, nil
}

// fetchGroupMetadata fetches group metadata with caching
func (c *UploadCommand) fetchGroupMetadata(ctx context.Context, groupID int) (*TorrentGroup, error) {
	cacheKey := fmt.Sprintf("group_%d", groupID)

	var cached TorrentGroup
	if c.Cache.LoadFrom(cacheKey, &cached, "redacted") {
		c.log("Using cached group metadata")
		return &cached, nil
	}

	meta, err := c.Client.GetTorrentGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Save to cache
	c.Cache.SaveTo(cacheKey, meta, "redacted")

	return meta, nil
}

// loadLocalTorrent loads metadata from the local torrent directory
func (c *UploadCommand) loadLocalTorrent() (*domain.Torrent, error) {
	// Try to load from extracted JSON files
	torrent := &domain.Torrent{
		RootPath: c.TorrentDir,
	}

	// Extract from FLAC files
	if err := c.extractFromFLACs(torrent); err != nil {
		return nil, err
	}

	return torrent, nil
}

// extractFromFLACs extracts metadata directly from FLAC files
func (c *UploadCommand) extractFromFLACs(torrent *domain.Torrent) error {
	var firstFileMetadata *tagging.Metadata

	// Walk directory to find FLAC files
	err := filepath.Walk(c.TorrentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(strings.ToLower(path), ".flac") {
			relPath, _ := filepath.Rel(c.TorrentDir, path)

			// Read metadata from FLAC file
			metadata, err := tagging.ReadMetadata(path)
			if err != nil {
				c.log("Warning: failed to read tags from %s: %v", relPath, err)
				return nil // Continue with other files
			}

			// Store first file's metadata for album-level info
			if firstFileMetadata == nil {
				firstFileMetadata = &metadata
			}

			// Convert to domain Track
			track, err := metadata.ToTrack(relPath)
			if err != nil {
				c.log("Warning: failed to convert metadata for %s: %v", relPath, err)
				return nil // Continue with other files
			}

			// Parse composers (may be comma-separated) - ToTrack only gets first one
			// Replace the single composer from ToTrack with all composers
			trackArtists := make([]domain.Artist, 0)
			if metadata.Composer != "" {
				composerArtists := domain.ParseArtistField(metadata.Composer)
				for _, comp := range composerArtists {
					trackArtists = append(trackArtists, domain.Artist{
						Name: comp.Name,
						Role: domain.RoleComposer,
					})
				}
			}

			// Parse performers (ARTIST field) - may be comma/semicolon separated
			if metadata.Artist != "" {
				performerArtists := domain.ParseArtistField(metadata.Artist)
				for _, perf := range performerArtists {
					// Default to ensemble role for ARTIST field performers
					// Roles will be inferred from context later if needed
					role := perf.Role
					if role == domain.RoleUnknown {
						role = domain.RoleEnsemble // Default for ARTIST field
					}
					trackArtists = append(trackArtists, domain.Artist{
						Name: perf.Name,
						Role: role,
					})
				}
			}

			// Update track with properly parsed artists
			if len(trackArtists) > 0 {
				track.Artists = trackArtists
			}

			torrent.Files = append(torrent.Files, track)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Extract album-level metadata from first file if not already set
	if firstFileMetadata != nil {
		if torrent.Title == "" && firstFileMetadata.Album != "" {
			torrent.Title = firstFileMetadata.Album
		}
		if torrent.OriginalYear == 0 && firstFileMetadata.Year != "" {
			if year, err := strconv.Atoi(firstFileMetadata.Year); err == nil && year > 0 {
				torrent.OriginalYear = year
			}
		}
		if len(torrent.AlbumArtist) == 0 && firstFileMetadata.AlbumArtist != "" {
			// Parse album artist field (may contain multiple artists, comma or semicolon separated)
			torrent.AlbumArtist = domain.ParseArtistField(firstFileMetadata.AlbumArtist)
		}
	}

	return nil
}

// creditsToArtists converts a slice of ArtistCredit to a slice of domain.Artist
func creditsToArtists(ac []ArtistCredit) []domain.Artist {
	var artists []domain.Artist
	for _, a := range ac {
		artists = append(artists, domain.Artist{
			Name: a.Name,
			Role: DomainRole(a.Role),
		})
	}
	return artists
}

// combineArtists combines all artist credits from group metadata and converts to domain.Artist
func (c *UploadCommand) combineArtists(group *TorrentGroup) []domain.Artist {
	var artists []domain.Artist

	artists = append(artists, creditsToArtists(group.Artists)...)
	artists = append(artists, creditsToArtists(group.Composers)...)
	artists = append(artists, creditsToArtists(group.Conductors)...)
	artists = append(artists, creditsToArtists(group.With)...)
	artists = append(artists, creditsToArtists(group.Producer)...)
	artists = append(artists, creditsToArtists(group.DJ)...)
	artists = append(artists, creditsToArtists(group.RemixedBy)...)

	return artists
}

// collectAllLocalArtists collects all artists from album and tracks (union)
func (c *UploadCommand) collectAllLocalArtists(torrent *domain.Torrent) map[domain.Artist]struct{} {
	artistMap := make(map[domain.Artist]struct{})

	// Add album artists
	for _, a := range torrent.AlbumArtist {
		artistMap[a] = struct{}{}
	}

	// Add track artists
	for _, fileLike := range torrent.Files {
		if track, ok := fileLike.(*domain.Track); ok {
			for _, a := range track.Artists {
				artistMap[a] = struct{}{}
			}
		}
	}

	return artistMap
}

// validateArtistsSuperset validates that local artists are a superset of Redacted artists
// Local can have additional artists, but must contain all Redacted artists
func (c *UploadCommand) validateArtistsSuperset(redacted []domain.Artist, local map[domain.Artist]struct{}) []error {
	var errors []error

	// Build a map of local artists by name for lookup
	localByName := make(map[string][]domain.Artist)
	for a := range local {
		localByName[a.Name] = append(localByName[a.Name], a)
	}

	// Check each Redacted artist exists in local
	for _, ra := range redacted {
		localArtists, exists := localByName[ra.Name]
		if !exists {
			errors = append(errors, fmt.Errorf("artist %q with role %q not found in local tags", ra.Name, ra.Role))
			continue
		}

		// Check if any local artist with this name has a compatible role
		found := false
		for _, localArtist := range localArtists {
			if c.rolesCompatible(ra.Role, localArtist.Role) {
				found = true
				break
			}
		}

		if !found {
			errors = append(errors, fmt.Errorf("artist %q with role %q not found in local tags (found with incompatible role)", ra.Name, ra.Role))
		}
	}

	// Note: We don't error on extra local artists - that's allowed (superset)
	return errors
}

// rolesCompatible checks if two roles are compatible (allows some flexibility)
func (c *UploadCommand) rolesCompatible(redactedRole, localRole domain.Role) bool {
	// Exact match
	if redactedRole == localRole {
		return true
	}

	// "artists" in Redacted (mapped to RolePerformer) can match ensemble, soloist, or performer in local
	// Redacted doesn't distinguish between these, so we allow flexible matching
	if redactedRole == domain.RolePerformer {
		if localRole == domain.RoleEnsemble || localRole == domain.RoleSoloist || localRole == domain.RolePerformer {
			return true
		}
	}

	return false
}

// mergeMetadata merges all metadata sources
// Uses local artists for upload (local is superset of Redacted)
func (c *UploadCommand) mergeMetadata(torrent *Torrent, _ *TorrentGroup, local *domain.Torrent, trumpReason string) *Metadata {
	// Collect all local artists (flat list)
	allLocalArtistsMap := c.collectAllLocalArtists(local)
	allLocalArtists := make([]domain.Artist, 0, len(allLocalArtistsMap))
	for a := range allLocalArtistsMap {
		allLocalArtists = append(allLocalArtists, a)
	}

	merged := &Metadata{
		// From local/extracted
		Title: local.Title,
		Year:  local.OriginalYear,

		Artists: allLocalArtists,

		// From Redacted torrent
		Format:    torrent.Format,
		Encoding:  torrent.Encoding,
		Media:     torrent.Media,
		Tags:      torrent.Tags,
		GroupID:   torrent.GroupID,
		TorrentID: torrent.TorrentID,

		// Remaster info
		Remastered:              torrent.Remastered,
		RemasterYear:            torrent.RemasterYear,
		RemasterTitle:           torrent.RemasterTitle,
		RemasterRecordLabel:     torrent.RemasterRecordLabel,
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
func (c *UploadCommand) generateTrumpReason(_ *domain.Torrent) string {
	// TODO: Analyze what was fixed based on validation results
	return "Corrected tags and filenames according to classical music guidelines"
}

// validateRequiredFields checks all required fields are present
func (c *UploadCommand) validateRequiredFields(meta *Metadata) error {
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
	if len(meta.Artists) == 0 {
		missing = append(missing, "artists")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

// prepareUploadRequest converts merged metadata to upload request
func (c *UploadCommand) prepareUploadRequest(meta *Metadata) *Upload {
	req := &Upload{
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
		Tags:               strings.Join(meta.Tags, ","),

		TrumpTorrent: meta.TorrentID,
		TrumpReason:  meta.TrumpReason,
	}

	// Convert artists to string arrays with importance values
	// All artists go in artists[] with appropriate importance values:
	// 1 = Main, 2 = Guest, 4 = Composer, 5 = Conductor, 7 = Producer
	for _, a := range meta.Artists {
		req.Artists = append(req.Artists, a.Name)
		req.Importance = append(req.Importance, RedactedImportance(a.Role))
	}

	return req
}

// createTorrentFile creates a .torrent file
func (c *UploadCommand) createTorrentFile(ctx context.Context, sourceDir string, announceURL string) (string, error) {
	// Check cache first
	torrentPath := filepath.Join(c.CacheDir, fmt.Sprintf("torrent_%d.torrent", c.TorrentID))
	if _, err := os.Stat(torrentPath); err == nil {
		c.log("Using cached torrent file")
		return torrentPath, nil
	}

	// Create torrent using mktorrent
	cmd := exec.CommandContext(ctx, "mktorrent",
		"-p",       // Private torrent
		"-l", "18", // Piece length 2^18 = 256KB
		"-a", announceURL, // Announce URL
		"-o", torrentPath, // Output file
		sourceDir, // Source directory
	)

	if c.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mktorrent failed: %w", err)
	}

	return torrentPath, nil
}

// printMergedMetadata prints metadata for dry run
func (c *UploadCommand) printMergedMetadata(meta *Metadata) {
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
	// Group by role for display
	byRole := make([][]string, domain.RoleUnknown+1)
	for _, a := range meta.Artists {
		byRole[int(a.Role)] = append(byRole[int(a.Role)], a.Name)
	}

	for role, artists := range byRole {
		if role == int(domain.RoleUnknown) || len(artists) == 0 {
			continue
		}
		fmt.Printf("\n%s:\n", domain.Role(role).String())
		for _, a := range artists {
			fmt.Printf("  - %s\n", a)
		}
	}

	// Display any unknown roles
	artists := byRole[domain.RoleUnknown]
	if len(artists) > 0 {
		fmt.Printf("\nUnknown:\n")
		for _, a := range artists {
			fmt.Printf("  - %s\n", a)
		}
	}

	fmt.Printf("\nTags: %s\n", strings.Join(meta.Tags, ", "))
	fmt.Printf("\nTrump Reason: %s\n", meta.TrumpReason)
	fmt.Printf("\nDescription:\n%s\n", meta.Description)
}

// log logs a message if verbose mode is enabled
func (c *UploadCommand) log(format string, args ...any) {
	if c.Verbose {
		fmt.Printf("[UPLOAD] "+format+"\n", args...)
	}
}
