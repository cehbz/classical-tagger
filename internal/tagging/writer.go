package tagging

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// FLACWriter writes FLAC metadata using the mewkiz/flac library.
// It preserves audio data bit-perfect while updating only metadata blocks.
type FLACWriter struct{}

// NewFLACWriter creates a new FLACWriter.
func NewFLACWriter() *FLACWriter {
	return &FLACWriter{}
}

// WriteTrack writes a track's metadata to a new FLAC file.
// The source file's audio data is preserved bit-perfect.
// The destination file is created in the output directory structure.
func (w *FLACWriter) WriteTrack(sourcePath, destPath string, track *domain.Track, album *domain.Album) error {
	// Parse source FLAC
	flacFile, err := flac.ParseFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to parse source FLAC: %w", err)
	}

	// Convert domain metadata to Vorbis comment tags
	tags := MetadataToVorbisComment(track, album)

	// Find or create VorbisComment block
	var cmtBlock *flacvorbis.MetaDataBlockVorbisComment
	var cmtIdx int = -1

	for idx, metaBlock := range flacFile.Meta {
		if metaBlock.Type == flac.VorbisComment {
			cmtBlock, err = flacvorbis.ParseFromMetaDataBlock(*metaBlock)
			if err != nil {
				return fmt.Errorf("failed to parse vorbis comment: %w", err)
			}
			cmtIdx = idx
			break
		}
	}

	if cmtBlock == nil {
		// Create new VorbisComment block
		cmtBlock = flacvorbis.New()
	}

	// Set vendor
	cmtBlock.Vendor = "classical-tagger"

	// Clear existing comments and add new ones
	cmtBlock.Comments = nil
	for key, value := range tags {
		cmtBlock.Add(strings.ToUpper(key), value)
	}

	// Marshal back to metadata block
	metaBlock := cmtBlock.Marshal()

	if cmtIdx >= 0 {
		// Replace existing
		flacFile.Meta[cmtIdx] = &metaBlock
	} else {
		// Insert after StreamInfo (index 0)
		if len(flacFile.Meta) > 0 {
			flacFile.Meta = append(flacFile.Meta[:1], append([]*flac.MetaDataBlock{&metaBlock}, flacFile.Meta[1:]...)...)
		} else {
			flacFile.Meta = append(flacFile.Meta, &metaBlock)
		}
	}

	// Write to destination
	if err := flacFile.Save(destPath); err != nil {
		return fmt.Errorf("failed to save FLAC: %w", err)
	}

	return nil
}

// MetadataToVorbisComment converts domain Track and Album to Vorbis comment tags.
// Returns a map of tag names to values following classical music conventions.
func MetadataToVorbisComment(track *domain.Track, album *domain.Album) map[string]string {
	tags := make(map[string]string)

	// Required tags per rules 2.3.16.4
	tags["TITLE"] = track.Title()
	tags["ALBUM"] = album.Title()
	tags["TRACKNUMBER"] = strconv.Itoa(track.Track())
	tags["DISCNUMBER"] = strconv.Itoa(track.Disc())

	// Find composer and format performers
	var composer *domain.Artist
	var performers []domain.Artist

	for _, artist := range track.Artists() {
		if artist.Role() == domain.RoleComposer {
			composer = &artist
		} else {
			performers = append(performers, artist)
		}
	}

	// COMPOSER tag (required for classical)
	if composer != nil {
		tags["COMPOSER"] = composer.Name()
	}

	// ARTIST tag (performers only, not composer)
	if len(performers) > 0 {
		tags["ARTIST"] = FormatArtists(performers)

		// Also add individual role-specific tags for classical music players
		for _, artist := range performers {
			switch artist.Role() {
			case domain.RoleSoloist:
				// Add to PERFORMER field (can be multiple)
				if existing, ok := tags["PERFORMER"]; ok {
					tags["PERFORMER"] = existing + "; " + artist.Name()
				} else {
					tags["PERFORMER"] = artist.Name()
				}
			case domain.RoleEnsemble:
				tags["ENSEMBLE"] = artist.Name()
			case domain.RoleConductor:
				tags["CONDUCTOR"] = artist.Name()
			}
		}
	}

	// Date fields following Vorbis/MusicBrainz conventions:
	// - ORIGINALDATE: Year of original recording/release
	// - DATE: Release date of this specific edition
	if album.OriginalYear() > 0 {
		tags["ORIGINALDATE"] = strconv.Itoa(album.OriginalYear())
	}

	// Edition information (if present)
	if edition := album.Edition(); edition != nil {
		// DATE: Edition year (this specific release)
		if edition.Year() > 0 {
			tags["DATE"] = strconv.Itoa(edition.Year())
		}
		if edition.Label() != "" {
			tags["LABEL"] = edition.Label()
		}
		if edition.CatalogNumber() != "" {
			tags["CATALOGNUMBER"] = edition.CatalogNumber()
		}
	}

	// ALBUMARTIST tag (if there are universal performers)
	albumArtist, _ := DetermineAlbumArtist(album)
	if albumArtist != "" {
		tags["ALBUMARTIST"] = albumArtist
	}

	return tags
}

// FormatArtists formats a list of artists according to classical music conventions.
// Format: "Soloist(s), Orchestra/Ensemble, Conductor"
// Composers are excluded from the ARTIST tag.
func FormatArtists(artists []domain.Artist) string {
	if len(artists) == 0 {
		return ""
	}

	var soloists []string
	var ensembles []string
	var conductors []string

	for _, artist := range artists {
		switch artist.Role() {
		case domain.RoleSoloist:
			soloists = append(soloists, artist.Name())
		case domain.RoleEnsemble:
			ensembles = append(ensembles, artist.Name())
		case domain.RoleConductor:
			conductors = append(conductors, artist.Name())
		case domain.RoleComposer:
			// Composers excluded from ARTIST tag
			continue
		}
	}

	// Build in order: soloists, ensembles, conductors
	var parts []string
	parts = append(parts, soloists...)
	parts = append(parts, ensembles...)
	parts = append(parts, conductors...)

	return strings.Join(parts, ", ")
}

// DetermineAlbumArtist finds performers that appear in all tracks.
// Returns the formatted album artist string and the list of universal artists.
// Per classical music guide: "When the performer(s) do not remain the same throughout
// all tracks, this tag is used to credit the one who does appear in all tracks."
func DetermineAlbumArtist(album *domain.Album) (string, []domain.Artist) {
	tracks := album.Tracks()
	if len(tracks) == 0 {
		return "", nil
	}

	// Build set of all non-composer artists from first track
	firstTrack := tracks[0]
	var candidates []domain.Artist
	for _, artist := range firstTrack.Artists() {
		if artist.Role() != domain.RoleComposer {
			candidates = append(candidates, artist)
		}
	}

	// Filter to only those appearing in ALL tracks
	var universal []domain.Artist
	for _, candidate := range candidates {
		appearsInAll := true
		for _, track := range tracks[1:] {
			found := false
			for _, artist := range track.Artists() {
				if artist.Name() == candidate.Name() && artist.Role() == candidate.Role() {
					found = true
					break
				}
			}
			if !found {
				appearsInAll = false
				break
			}
		}
		if appearsInAll {
			universal = append(universal, candidate)
		}
	}

	if len(universal) == 0 {
		return "", nil
	}

	return FormatArtists(universal), universal
}
