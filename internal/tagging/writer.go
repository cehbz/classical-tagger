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
	tags["TITLE"] = track.Title
	tags["ALBUM"] = album.Title
	tags["TRACKNUMBER"] = strconv.Itoa(track.Track)
	tags["DISCNUMBER"] = strconv.Itoa(track.Disc)

	// Find composer and format performers
	var composer *domain.Artist
	var performers []domain.Artist

	for _, artist := range track.Artists {
		if artist.Role == domain.RoleComposer {
			composer = &artist
		} else {
			// Preserve incoming order; grouping is handled by FormatArtists which now appends Unknown last
			performers = append(performers, artist)
		}
	}

	// COMPOSER tag (required for classical)
	if composer != nil {
		tags["COMPOSER"] = composer.Name
	}

	// ARTIST tag (performers only, not composer)
	if len(performers) > 0 {
		tags["ARTIST"] = domain.FormatArtists(performers)

		// Also add individual role-specific tags for classical music players
		for _, artist := range performers {
			switch artist.Role {
			case domain.RoleSoloist:
				// Add to PERFORMER field (can be multiple)
				if existing, ok := tags["PERFORMER"]; ok {
					tags["PERFORMER"] = existing + "; " + artist.Name
				} else {
					tags["PERFORMER"] = artist.Name
				}
			case domain.RoleEnsemble:
				tags["ENSEMBLE"] = artist.Name
			case domain.RoleConductor:
				tags["CONDUCTOR"] = artist.Name
			}
		}
	}

	// Date fields following Vorbis/MusicBrainz conventions:
	// - ORIGINALDATE: Year of original recording/release
	// - DATE: Release date of this specific edition
	if album.OriginalYear > 0 {
		tags["ORIGINALDATE"] = strconv.Itoa(album.OriginalYear)
	}

	// Edition information (if present)
	if edition := album.Edition; edition != nil {
		// DATE: Edition year (this specific release)
		if edition.Year > 0 {
			tags["DATE"] = strconv.Itoa(edition.Year)
		}
		if edition.Label != "" {
			tags["LABEL"] = edition.Label
		}
		if edition.CatalogNumber != "" {
			tags["CATALOGNUMBER"] = edition.CatalogNumber
		}
	}

	// ALBUMARTIST tag (if set in album)
	if len(album.AlbumArtist) > 0 {
		tags["ALBUMARTIST"] = domain.FormatArtists(album.AlbumArtist)
	}

	return tags
}
