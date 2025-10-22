package tagging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/go-flac/go-flac/v2"
	"github.com/go-flac/flacvorbis/v2"
)

// FLACWriter writes tags to FLAC files with data loss protection.
type FLACWriter struct {
	dryRun bool
}

// NewFLACWriter creates a new FLAC writer.
func NewFLACWriter() *FLACWriter {
	return &FLACWriter{
		dryRun: false,
	}
}

// SetDryRun enables or disables dry-run mode.
// In dry-run mode, no files are actually modified.
func (w *FLACWriter) SetDryRun(dryRun bool) {
	w.dryRun = dryRun
}

// TagComparison shows what changed in a tag.
type TagComparison struct {
	Field     string
	OldValue  string
	NewValue  string
	Status    string // "unchanged", "updated", "added", "preserved", "would-lose-data"
}

// WriteTrack writes track and album metadata to a FLAC file.
// Preserves all existing tags, only adds/updates music metadata.
// Returns error if writing would cause data loss.
func (w *FLACWriter) WriteTrack(path string, track *domain.Track, album *domain.Album) error {
	if w.dryRun {
		fmt.Printf("[DRY RUN] Would write track: %s\n", path)
		return nil
	}

	// Compare before writing to detect data loss
	comparisons, err := w.compareBeforeWrite(path, track, album)
	if err != nil {
		return fmt.Errorf("pre-write comparison failed: %w", err)
	}

	// Check for data loss
	for _, cmp := range comparisons {
		if cmp.Status == "would-lose-data" {
			return fmt.Errorf("would lose data in %s: old=%q new=%q",
				cmp.Field, cmp.OldValue, cmp.NewValue)
		}
	}

	// Write the tags
	return w.writeTags(path, track, album)
}

// WriteTrackWithReport writes tags and returns comparison report.
func (w *FLACWriter) WriteTrackWithReport(path string, track *domain.Track, album *domain.Album) ([]TagComparison, error) {
	comparisons, err := w.compareBeforeWrite(path, track, album)
	if err != nil {
		return nil, err
	}

	if !w.dryRun {
		err = w.writeTags(path, track, album)
	}

	return comparisons, err
}

// BackupFile creates a backup copy of a file.
// Returns the path to the backup file.
func (w *FLACWriter) BackupFile(path string) (string, error) {
	// Generate backup path with timestamp
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup-%s%s", base, timestamp, ext)

	// Copy file
	src, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("failed to copy data: %w", err)
	}

	return backupPath, nil
}

// RestoreBackup restores a file from a backup.
func (w *FLACWriter) RestoreBackup(backupPath, originalPath string) error {
	// Copy backup back to original
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(originalPath)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// ============================================================================
// Internal Implementation
// ============================================================================

// tagData is internal struct for marshaling domain to FLAC tags.
type tagData struct {
	title         string
	composer      string
	artist        string
	album         string
	albumArtist   string
	date          string
	originalDate  string
	trackNumber   string
	discNumber    string
	label         string
	catalogNumber string
}

// trackToTagData converts Track + Album to tagData.
func trackToTagData(track *domain.Track, album *domain.Album) tagData {
	tags := tagData{
		title:       track.Title(),
		composer:    track.Composer().Name(),
		artist:      formatArtists(track.Artists()),
		album:       album.Title(),
		trackNumber: strconv.Itoa(track.Track()),
		discNumber:  strconv.Itoa(track.Disc()),
	}

	// Determine album artist (universal performers only)
	albumArtist, universal := determineAlbumArtist(album)
	tags.albumArtist = albumArtist
	if len(universal) > 1 {
		// Log for information
		fmt.Printf("INFO: Multiple universal performers found: %v\n", universal)
	}

	// Set date fields
	if album.Edition() != nil {
		// Edition year goes in DATE
		tags.date = strconv.Itoa(album.Edition().Year())
		tags.originalDate = strconv.Itoa(album.OriginalYear())

		// Edition metadata
		tags.label = album.Edition().Label()
		tags.catalogNumber = album.Edition().CatalogNumber()
	} else {
		// No edition - use original year for both
		tags.date = strconv.Itoa(album.OriginalYear())
		tags.originalDate = strconv.Itoa(album.OriginalYear())
	}

	return tags
}

// formatArtists formats artists following guide rules:
// "in that order (soloists, then ensembles, then conductors)"
func formatArtists(artists []domain.Artist) string {
	// Group by role
	var soloists []string
	var ensembles []string
	var conductors []string
	var others []string

	for _, artist := range artists {
		switch artist.Role() {
		case domain.RoleComposer:
			// Skip - composer has own tag
			continue
		case domain.RoleSoloist:
			soloists = append(soloists, artist.Name())
		case domain.RoleEnsemble:
			ensembles = append(ensembles, artist.Name())
		case domain.RoleConductor:
			conductors = append(conductors, artist.Name())
		default:
			// Guest, Arranger, etc.
			others = append(others, artist.Name())
		}
	}

	// Combine in order: soloists, ensembles, conductors, others
	var parts []string
	parts = append(parts, soloists...)
	parts = append(parts, ensembles...)
	parts = append(parts, conductors...)
	parts = append(parts, others...)

	return strings.Join(parts, ", ")
}

// determineAlbumArtist returns album artist tag value.
// Only set if performers appear on ALL tracks.
// Returns formatted string and list of universal artists.
func determineAlbumArtist(album *domain.Album) (string, []string) {
	if len(album.Tracks()) == 0 {
		return "", nil
	}

	// Count how many tracks each artist appears on
	artistCounts := make(map[string]int)
	artistRoles := make(map[string]domain.Role) // Track role for ordering

	for _, track := range album.Tracks() {
		for _, artist := range track.Artists() {
			if artist.Role() != domain.RoleComposer {
				artistCounts[artist.Name()]++
				artistRoles[artist.Name()] = artist.Role()
			}
		}
	}

	// Find artists appearing on ALL tracks
	totalTracks := len(album.Tracks())
	var universalArtists []string
	for artistName, count := range artistCounts {
		if count == totalTracks {
			universalArtists = append(universalArtists, artistName)
		}
	}

	if len(universalArtists) == 0 {
		return "", nil
	}

	// Sort by role priority: Soloist → Ensemble → Conductor → Others
	sortArtistsByRole(universalArtists, artistRoles)

	formatted := strings.Join(universalArtists, ", ")
	return formatted, universalArtists
}

// sortArtistsByRole sorts artists by role priority in place.
func sortArtistsByRole(artists []string, roles map[string]domain.Role) {
	// Assign priority to each role
	rolePriority := func(role domain.Role) int {
		switch role {
		case domain.RoleSoloist:
			return 0
		case domain.RoleEnsemble:
			return 1
		case domain.RoleConductor:
			return 2
		default:
			return 3
		}
	}

	// Simple bubble sort by role priority
	for i := 0; i < len(artists); i++ {
		for j := i + 1; j < len(artists); j++ {
			priorityI := rolePriority(roles[artists[i]])
			priorityJ := rolePriority(roles[artists[j]])
			if priorityI > priorityJ {
				artists[i], artists[j] = artists[j], artists[i]
			}
		}
	}
}

// splitArtists splits comma/semicolon separated artist string.
func splitArtists(s string) []string {
	parts := regexp.MustCompile(`[,;]`).Split(s, -1)
	var artists []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			artists = append(artists, trimmed)
		}
	}
	return artists
}

// isArtistSuperset checks if new artist list contains all old artists.
// Returns (ok, reason).
func isArtistSuperset(newVal, oldVal string) (bool, string) {
	if oldVal == "" {
		return true, "" // No old value, nothing to lose
	}

	oldArtists := splitArtists(oldVal)
	newArtists := splitArtists(newVal)

	// Check each old artist is in new list
	for _, oldArtist := range oldArtists {
		found := false
		oldNorm := strings.TrimSpace(strings.ToLower(oldArtist))

		for _, newArtist := range newArtists {
			newNorm := strings.TrimSpace(strings.ToLower(newArtist))
			if oldNorm == newNorm {
				found = true
				break
			}
		}

		if !found {
			return false, fmt.Sprintf("would lose artist %q", oldArtist)
		}
	}

	return true, ""
}

// compareBeforeWrite compares existing vs proposed tags.
func (w *FLACWriter) compareBeforeWrite(path string, track *domain.Track, album *domain.Album) ([]TagComparison, error) {
	// Read existing tags
	existing, err := w.readExistingTags(path)
	if err != nil {
		// If we can't read, assume empty (new file)
		existing = make(map[string]string)
	}

	// Build proposed tags
	proposed := w.buildProposedTags(track, album)

	var comparisons []TagComparison

	// Check all fields (existing + proposed)
	allFields := make(map[string]bool)
	for field := range existing {
		allFields[field] = true
	}
	for field := range proposed {
		allFields[field] = true
	}

	for field := range allFields {
		oldValue := existing[field]
		newValue := proposed[field]

		if oldValue == "" && newValue != "" {
			// Adding new tag
			comparisons = append(comparisons, TagComparison{
				Field:    field,
				OldValue: "(not set)",
				NewValue: newValue,
				Status:   "added",
			})
		} else if oldValue != "" && newValue == "" {
			// Preserving existing tag (we're not setting it)
			comparisons = append(comparisons, TagComparison{
				Field:    field,
				OldValue: oldValue,
				NewValue: oldValue,
				Status:   "preserved",
			})
		} else if oldValue == newValue {
			// Unchanged
			comparisons = append(comparisons, TagComparison{
				Field:    field,
				OldValue: oldValue,
				NewValue: newValue,
				Status:   "unchanged",
			})
		} else {
			// Updated - verify no data loss
			var ok bool

			if field == "ARTIST" {
				ok, _ = isArtistSuperset(newValue, oldValue)
				if !ok {
					comparisons = append(comparisons, TagComparison{
						Field:    field,
						OldValue: oldValue,
						NewValue: newValue,
						Status:   "would-lose-data",
					})
				}
			} else {
				// For other fields, simple containment check
				ok = strings.Contains(newValue, oldValue)
				if !ok {
					comparisons = append(comparisons, TagComparison{
						Field:    field,
						OldValue: oldValue,
						NewValue: newValue,
						Status:   "would-lose-data",
					})
				}
			}

			if !ok {
				comparisons = append(comparisons, TagComparison{
					Field:    field,
					OldValue: oldValue,
					NewValue: newValue,
					Status:   "would-lose-data",
				})
			} else {
				comparisons = append(comparisons, TagComparison{
					Field:    field,
					OldValue: oldValue,
					NewValue: newValue,
					Status:   "updated",
				})
			}
		}
	}

	return comparisons, nil
}

// readExistingTags reads current tags from FLAC file.
func (w *FLACWriter) readExistingTags(path string) (map[string]string, error) {
	f, err := flac.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FLAC: %w", err)
	}

	// Find vorbis comment block
	for _, meta := range f.Meta {
		if meta.Type == flac.VorbisComment {
			cmt, err := flacvorbis.ParseFromMetaDataBlock(*meta)
			if err != nil {
				continue
			}

			// Extract all tags
			tags := make(map[string]string)
			for _, comment := range cmt.Comments {
				parts := strings.SplitN(comment, "=", 2)
				if len(parts) == 2 {
					field := strings.ToUpper(strings.TrimSpace(parts[0]))
					value := strings.TrimSpace(parts[1])
					tags[field] = value
				}
			}
			return tags, nil
		}
	}

	// No vorbis comment block found
	return make(map[string]string), nil
}

// buildProposedTags builds map of tags we want to write.
func (w *FLACWriter) buildProposedTags(track *domain.Track, album *domain.Album) map[string]string {
	tags := trackToTagData(track, album)

	proposed := make(map[string]string)

	if tags.title != "" {
		proposed["TITLE"] = tags.title
	}
	if tags.composer != "" {
		proposed["COMPOSER"] = tags.composer
	}
	if tags.artist != "" {
		proposed["ARTIST"] = tags.artist
	}
	if tags.album != "" {
		proposed["ALBUM"] = tags.album
	}
	if tags.albumArtist != "" {
		proposed["ALBUMARTIST"] = tags.albumArtist
	}
	if tags.date != "" {
		proposed["DATE"] = tags.date
	}
	if tags.originalDate != "" {
		proposed["ORIGINALDATE"] = tags.originalDate
	}
	if tags.trackNumber != "" {
		proposed["TRACKNUMBER"] = tags.trackNumber
	}
	if tags.discNumber != "" {
		proposed["DISCNUMBER"] = tags.discNumber
	}
	if tags.label != "" {
		proposed["LABEL"] = tags.label
	}
	if tags.catalogNumber != "" {
		proposed["CATALOGNUMBER"] = tags.catalogNumber
	}

	return proposed
}

// writeTags writes tags to FLAC file using go-flac.
func (w *FLACWriter) writeTags(path string, track *domain.Track, album *domain.Album) error {
	// 1. Parse FLAC file
	f, err := flac.ParseFile(path)
	if err != nil {
		return fmt.Errorf("failed to parse FLAC: %w", err)
	}

	// 2. Read existing tags
	existingTags, _ := w.readExistingTags(path)

	// 3. Build proposed tags
	proposedTags := w.buildProposedTags(track, album)

	// 4. Merge: preserve existing + overlay new
	mergedTags := make(map[string]string)
	for k, v := range existingTags {
		mergedTags[k] = v
	}
	for k, v := range proposedTags {
		mergedTags[k] = v
	}

	// 5. Find or create vorbis comment block
	cmt, cmtIdx := w.extractVorbisComment(f)
	if cmt == nil {
		cmt = flacvorbis.New()
		cmtIdx = -1
	}

	// 6. Clear and set all tags
	cmt = flacvorbis.New()
	for field, value := range mergedTags {
		cmt.Add(field, value)
	}

	// 7. Marshal back to metadata block
	cmtsmeta := cmt.Marshal()
	if cmtIdx >= 0 {
		f.Meta[cmtIdx] = &cmtsmeta
	} else {
		f.Meta = append(f.Meta, &cmtsmeta)
	}

	// 8. Save to temp file (CRITICAL: not same as input!)
	tempPath := path + ".tmp"
	err = f.Save(tempPath)
	if err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	// 9. Atomic rename
	err = os.Rename(tempPath, path)
	if err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

// extractVorbisComment finds vorbis comment block in FLAC file.
func (w *FLACWriter) extractVorbisComment(f *flac.File) (*flacvorbis.MetaDataBlockVorbisComment, int) {
	for idx, meta := range f.Meta {
		if meta.Type == flac.VorbisComment {
			cmt, err := flacvorbis.ParseFromMetaDataBlock(*meta)
			if err == nil {
				return cmt, idx
			}
		}
	}
	return nil, -1
}
