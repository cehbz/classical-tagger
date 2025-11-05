package validation

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TorrentBuilder provides a fluent API for building domain.Torrent instances.
type TorrentBuilder struct {
	torrent *domain.Torrent
}

// TrackBuilder provides a fluent API for building domain.Track instances.
type TrackBuilder struct {
	track   *domain.Track
	builder *TorrentBuilder // reference to parent builder for Build() to return
}

// NewTorrent creates a new TorrentBuilder with sensible defaults.
// Defaults: Title="Torrent", OriginalYear=1963, single track (Disc=1, Track=1, Title="Symphony")
// with Composer="Beethoven" and Ensemble="Orchestra".
func NewTorrent() *TorrentBuilder {
	return &TorrentBuilder{
		torrent: &domain.Torrent{
			Title:        "Torrent",
			OriginalYear: 0,
			Files: []domain.FileLike{
				&domain.Track{
					File: domain.File{
						Path: "01.flac",
					},
					Disc:  1,
					Track: 1,
					Title: "Symphony",
					Artists: []domain.Artist{
						{Name: "Beethoven", Role: domain.RoleComposer},
						{Name: "Orchestra", Role: domain.RoleEnsemble},
					},
				},
			},
		},
	}
}

// WithTitle sets the torrent title.
func (b *TorrentBuilder) WithTitle(title string) *TorrentBuilder {
	b.torrent.Title = title
	return b
}

// WithOriginalYear sets the torrent's original year.
func (b *TorrentBuilder) WithOriginalYear(year int) *TorrentBuilder {
	b.torrent.OriginalYear = year
	return b
}

// WithComposer adds a composer to the default track (first track).
// If no tracks exist, creates a default track first.
func (b *TorrentBuilder) WithComposer(name string) *TorrentBuilder {
	return b.WithArtists(domain.Artist{
		Name: name,
		Role: domain.RoleComposer,
	})
}

// WithComposers adds multiple composers to the default track (variadic).
func (b *TorrentBuilder) WithComposers(names ...string) *TorrentBuilder {
	for _, name := range names {
		b.WithComposer(name)
	}
	return b
}

// WithArtist adds a specific artist to the default track.
func (b *TorrentBuilder) WithArtist(name string, role domain.Role) *TorrentBuilder {
	return b.WithArtists(domain.Artist{Name: name, Role: role})
}

// WithArtists adds multiple artists to the default track (variadic).
func (b *TorrentBuilder) WithArtists(artists ...domain.Artist) *TorrentBuilder {
	b.ensureDefaultTrack()
	files := b.torrent.Files
	if len(files) > 0 {
		if track, ok := files[0].(*domain.Track); ok {
			track.Artists = append(track.Artists, artists...)
		}
	}
	return b
}

// WithEdition sets the torrent edition.
func (b *TorrentBuilder) WithEdition(label, catalogNumber string, year int) *TorrentBuilder {
	b.torrent.Edition = &domain.Edition{
		Label:         label,
		CatalogNumber: catalogNumber,
		Year:          year,
	}
	return b
}

// WithoutEdition explicitly removes the edition.
func (b *TorrentBuilder) WithoutEdition() *TorrentBuilder {
	b.torrent.Edition = nil
	return b
}

// AddTrack returns a TrackBuilder for adding a new track to the torrent.
func (b *TorrentBuilder) AddTrack() *TrackBuilder {
	return &TrackBuilder{
		track: &domain.Track{
			File: domain.File{
				Path: "",
			},
			Disc:  1,
			Track: len(b.torrent.Tracks()) + 1,
			Artists: []domain.Artist{
				{Name: "Beethoven", Role: domain.RoleComposer},
				{Name: "Orchestra", Role: domain.RoleEnsemble},
			},
		},
		builder: b,
	}
}

// AddFiles adds multiple files to the torrent (variadic)
func (b *TorrentBuilder) AddTracks(tracks ...domain.FileLike) *TorrentBuilder {
	b.torrent.Files = append(b.torrent.Files, tracks...)
	return b
}

// WithTracks replaces all tracks in the torrent.
func (b *TorrentBuilder) WithTracks(tracks []domain.FileLike) *TorrentBuilder {
	b.torrent.Files = tracks
	return b
}

// ClearTracks removes all tracks from the torrent.
func (b *TorrentBuilder) ClearTracks() *TorrentBuilder {
	b.torrent.Files = nil
	return b
}

// Build returns the constructed torrent (converted from torrent).
func (b *TorrentBuilder) Build() *domain.Torrent {
	// Convert Torrent to Torrent
	return b.torrent
}

// ensureDefaultTrack ensures there's at least one track with defaults.
func (b *TorrentBuilder) ensureDefaultTrack() {
	if len(b.torrent.Tracks()) == 0 {
		b.torrent.Files = []domain.FileLike{
			&domain.Track{
				File: domain.File{
					Path: "01.flac",
				},
				Disc:  1,
				Track: 1,
				Title: "Symphony",
				Artists: []domain.Artist{
					{Name: "Beethoven", Role: domain.RoleComposer},
					{Name: "Orchestra", Role: domain.RoleEnsemble},
				},
			},
		}
	}
}

// WithDisc sets the disc number for the track.
func (tb *TrackBuilder) WithDisc(disc int) *TrackBuilder {
	tb.track.Disc = disc
	return tb
}

// WithTrack sets the track number.
func (tb *TrackBuilder) WithTrack(track int) *TrackBuilder {
	tb.track.Track = track
	return tb
}

// WithTitle sets the track title.
func (tb *TrackBuilder) WithTitle(title string) *TrackBuilder {
	tb.track.Title = title
	return tb
}

// WithFilename sets the track filename.
func (tb *TrackBuilder) WithFilename(filename string) *TrackBuilder {
	tb.track.File.Path = filename
	return tb
}

// WithArtist adds a specific artist to the track.
func (tb *TrackBuilder) WithArtist(name string, role domain.Role) *TrackBuilder {
	tb.track.Artists = append(tb.track.Artists, domain.Artist{
		Name: name,
		Role: role,
	})
	return tb
}

// WithArtists adds multiple artists to the track (variadic).
func (tb *TrackBuilder) WithArtists(artists ...domain.Artist) *TrackBuilder {
	tb.track.Artists = append(tb.track.Artists, artists...)
	return tb
}

// ClearArtists removes all artists from the track.
func (tb *TrackBuilder) ClearArtists() *TrackBuilder {
	tb.track.Artists = nil
	return tb
}

// Build adds the track to the torrent and returns the torrent builder.
func (tb *TrackBuilder) Build() *TorrentBuilder {
	tb.builder.torrent.Files = append(tb.builder.torrent.Files, tb.track)
	return tb.builder
}

// lastNames extracts last name(s) from a composer name
// "Ludwig van Beethoven" -> ["Beethoven"]
// "J.S. Bach" -> ["Bach"]
// "Johann Sebastian Bach" -> ["Bach"]
// "Beethoven, Ludwig van" -> ["Beethoven"]
func lastName(name string) string {
	// Handle reversed format: "Beethoven, Ludwig van"
	if strings.Contains(name, ",") {
		parts := strings.Split(name, ",")
		return strings.TrimSpace(parts[0])
	}

	// Handle normal format: "Ludwig van Beethoven"
	parts := strings.Fields(name)
	if len(parts) == 0 {
		// if it's only one word, assume that's the last name
		return name
	}

	// If the penultimate part is a lowercase particle, include it with the last word
	if len(parts) >= 2 {
		penult := parts[len(parts)-2]
		last := parts[len(parts)-1]
		if penult == strings.ToLower(penult) {
			return penult + " " + last
		}
	}
	// Otherwise, use the last word
	return parts[len(parts)-1]
}

// titleFile pairs track title with filename
type titleFile struct {
	Title    string
	Filename string
}

// buildTorrentWithTitlesAndFilenames creates torrent with specific title/filename pairs
func buildTorrentWithTitlesAndFilenames(titleFiles []titleFile) *domain.Torrent {
	builder := NewTorrent().
		WithTitle("Test Torrent").
		ClearTracks()

	for i, tf := range titleFiles {
		builder.AddTrack().
			WithTrack(i + 1).
			WithTitle(tf.Title).
			WithFilename(tf.Filename).
			Build()
	}

	return builder.Build()
}

// isAudioFile checks if a filename is an audio file based on extension
func isAudioFile(filename string) bool {
	audioExtensions := []string{".flac", ".mp3", ".wav", ".m4a", ".aac", ".ogg", ".wma", ".ape"}
	filenameLower := strings.ToLower(filename)
	for _, ext := range audioExtensions {
		if strings.HasSuffix(filenameLower, ext) {
			return true
		}
	}
	return false
}

// Helper function to build a torrent with specific filenames
func buildTorrentWithFilenames(filenames ...string) *domain.Torrent {
	builder := NewTorrent().
		WithTitle("Beethoven Symphonies").
		ClearTracks()

	trackNum := 1
	for _, filename := range filenames {
		if isAudioFile(filename) {
			// Audio files become Track objects
			builder.AddTrack().
				WithTrack(trackNum).
				WithTitle("Symphony No. 5").
				WithFilename(filename).
				ClearArtists().
				WithArtists(
					domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer},
					domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},
					domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor},
				).
				Build()
			trackNum++
		} else {
			// Non-audio files become File objects
			builder.AddTracks(&domain.File{Path: filename})
		}
	}

	return builder.Build()
}

// buildSingleComposerTorrentWithFilenames creates torrent with one composer
func buildSingleComposerTorrentWithFilenames(composerName string, filenames ...string) *domain.Torrent {
	builder := NewTorrent().ClearTracks()

	for i := range filenames {
		builder.AddTrack().
			WithTrack(i+1).
			WithTitle("Work "+string(rune('A'+i))).
			WithFilename(filenames[i]).
			ClearArtists().
			WithArtists(
				domain.Artist{Name: composerName, Role: domain.RoleComposer},
				domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
			).
			Build()
	}

	return builder.Build()
}

// composerTrack represents a track with specific composer
type composerTrack struct {
	ComposerName string
	TrackNum     int
	Filename     string
}

// buildMultiComposerTorrentWithFilenames creates torrent with multiple composers
func buildMultiComposerTorrentWithFilenames(composerTracks ...[]composerTrack) *domain.Torrent {
	builder := NewTorrent().
		WithTitle("Various Composers").
		ClearTracks()

	for _, ctList := range composerTracks {
		for _, ct := range ctList {
			builder.AddTrack().
				WithTrack(ct.TrackNum).
				WithTitle("Work "+string(rune('A'+ct.TrackNum))).
				WithFilename(ct.Filename).
				ClearArtists().
				WithArtists(
					domain.Artist{Name: ct.ComposerName, Role: domain.RoleComposer},
					domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
				).
				Build()
		}
	}

	return builder.Build()
}

// trackFile pairs track number with filename
type trackFile struct {
	TrackNum int
	Filename string
}

// buildTorrentWithTrackFilenames creates a torrent with specific track/filename pairs
func buildTorrentWithTrackFilenames(trackFiles ...trackFile) *domain.Torrent {
	builder := NewTorrent().
		WithTitle("Test Torrent").
		ClearTracks()

	for _, tf := range trackFiles {
		builder.AddTrack().
			WithTrack(tf.TrackNum).
			WithTitle("Work " + string(rune('A'+tf.TrackNum))).
			WithFilename(tf.Filename).
			Build()
	}

	return builder.Build()
}

// buildMultiDiscTorrentWithFilenames creates multi-disc torrent
func buildMultiDiscTorrentWithFilenames(disc1, disc2 []trackFile) *domain.Torrent {
	builder := NewTorrent().
		WithTitle("Multi-Disc Torrent").
		ClearTracks()

	for _, tf := range disc1 {
		builder.AddTrack().
			WithDisc(1).
			WithTrack(tf.TrackNum).
			WithTitle("Work " + string(rune('A'+tf.TrackNum))).
			WithFilename(tf.Filename).
			Build()
	}

	for _, tf := range disc2 {
		builder.AddTrack().
			WithDisc(2).
			WithTrack(tf.TrackNum).
			WithTitle("Work " + string(rune('A'+tf.TrackNum))).
			WithFilename(tf.Filename).
			Build()
	}

	return builder.Build()
}

// discTrack represents a track with disc and track number
type discTrack struct {
	Disc     int
	TrackNum int
}

// buildTorrentWithDiscTracks creates a torrent with specific disc/track combinations
func buildTorrentWithDiscTracks(discTracks []discTrack) *domain.Torrent {
	builder := NewTorrent().
		WithTitle("Multi-Disc Torrent").
		ClearTracks()

	for _, dt := range discTracks {
		builder.AddTrack().
			WithDisc(dt.Disc).
			WithTrack(dt.TrackNum).
			WithTitle(fmt.Sprintf("Track D%d-T%d", dt.Disc, dt.TrackNum)).
			WithFilename(fmt.Sprintf("CD%d/%02d - Track.flac", dt.Disc, dt.TrackNum)).
			Build()
	}

	return builder.Build()
}

// buildTorrentWithArtistName creates torrent with specific artist name
func buildTorrentWithArtistName(artistName string) *domain.Torrent {
	return NewTorrent().
		WithComposer(artistName).
		Build()
}

// buildTorrentWithTrackTitle creates torrent with specific track title
func buildTorrentWithTrackTitle(title string) *domain.Torrent {
	return NewTorrent().
		ClearTracks().
		AddTrack().
		WithTitle(title).
		Build().
		Build()
}

// buildTorrentWithFilenamesAndDiscs creates torrent with specific filenames and disc numbers
func buildTorrentWithFilenamesAndDiscs(filenames []string, discs []int) *domain.Torrent {
	builder := NewTorrent().ClearTracks()

	for i, filename := range filenames {
		builder.AddTrack().
			WithDisc(discs[i]).
			WithTrack(i + 1).
			WithTitle("Track").
			WithFilename(filename).
			Build()
	}

	return builder.Build()
}

// buildTorrentWithCompleteEdition creates torrent with complete edition information
func buildTorrentWithCompleteEdition() *domain.Torrent {
	return NewTorrent().
		WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).
		Build()
}

// buildTorrentWithConsistentSoloist creates torrent with same soloist on all tracks
func buildTorrentWithConsistentSoloist(soloistName string, trackCount int) *domain.Torrent {
	builder := NewTorrent().ClearTracks()

	for i := 0; i < trackCount; i++ {
		builder.AddTrack().
			WithTrack(i+1).
			WithTitle(fmt.Sprintf("Concerto No. %d", i+1)).
			ClearArtists().
			WithArtists(
				domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
				domain.Artist{Name: soloistName, Role: domain.RoleSoloist},
				domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
			).
			Build()
	}

	return builder.Build()
}

// buildTorrentWithGuestSoloist creates torrent where one soloist appears infrequently
func buildTorrentWithGuestSoloist(mainSoloist, guestSoloist string, trackCount int) *domain.Torrent {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	main := domain.Artist{Name: mainSoloist, Role: domain.RoleSoloist}
	guest := domain.Artist{Name: guestSoloist, Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

	builder := NewTorrent().ClearTracks()

	for i := 0; i < trackCount; i++ {
		var artists []domain.Artist
		// Guest appears only on first track
		if i == 0 {
			artists = []domain.Artist{composer, guest, ensemble}
		} else {
			artists = []domain.Artist{composer, main, ensemble}
		}

		builder.AddTrack().
			WithTrack(i + 1).
			WithTitle(fmt.Sprintf("Concerto No. %d", i+1)).
			ClearArtists().
			WithArtists(artists...).
			Build()
	}

	return builder.Build()
}

// buildTorrentWithGuestInTitle creates torrent with guest indicated in title
func buildTorrentWithGuestInTitle(mainSoloist, guestSoloist string, trackCount int) *domain.Torrent {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	main := domain.Artist{Name: mainSoloist, Role: domain.RoleSoloist}
	guest := domain.Artist{Name: guestSoloist, Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

	builder := NewTorrent().ClearTracks()

	for i := 0; i < trackCount; i++ {
		var artists []domain.Artist
		var title string

		// Guest appears only on first track, indicated in title
		if i == 0 {
			artists = []domain.Artist{composer, guest, ensemble}
			title = fmt.Sprintf("Concerto No. %d (feat. %s)", i+1, guestSoloist)
		} else {
			artists = []domain.Artist{composer, main, ensemble}
			title = fmt.Sprintf("Concerto No. %d", i+1)
		}

		builder.AddTrack().
			WithTrack(i + 1).
			WithTitle(title).
			ClearArtists().
			WithArtists(artists...).
			Build()
	}

	return builder.Build()
}

// buildTorrentWithGoodCaps creates torrent with proper capitalization
func buildTorrentWithGoodCaps() *domain.Torrent {
	builder := NewTorrent().
		WithTitle("Beethoven - Symphonies [1963] [FLAC]").
		ClearTracks()

	for i := 0; i < 3; i++ {
		builder.AddTrack().
			WithTrack(i+1).
			WithTitle(fmt.Sprintf("Symphony No. %d", i+1)).
			ClearArtists().
			WithArtists(
				domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer},
				domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			).
			Build()
	}

	return builder.Build()
}

// normalizeNameForInclusion normalizes a name for inclusion checks:
// - lowercases letters
// - strips spaces and punctuation (keeps letters with diacritics and digits)
func normalizeNameForInclusion(s string) string {
	var out []rune
	for _, r := range s {
		lr := unicode.ToLower(r)
		if unicode.IsLetter(lr) || unicode.IsDigit(lr) {
			out = append(out, lr)
		}
	}
	return string(out)
}

// buildTorrentWithBadCaps creates torrent with poor capitalization
func buildTorrentWithBadCaps() *domain.Torrent {
	builder := NewTorrent().
		WithTitle("BEETHOVEN - SYMPHONIES").
		ClearTracks()

	for i := 0; i < 3; i++ {
		builder.AddTrack().
			WithTrack(i+1).
			WithTitle(fmt.Sprintf("SYMPHONY NO. %d", i+1)).
			ClearArtists().
			WithArtists(
				domain.Artist{Name: "BEETHOVEN", Role: domain.RoleComposer},
				domain.Artist{Name: "berlin philharmonic", Role: domain.RoleEnsemble},
			).
			Build()
	}

	return builder.Build()
}
