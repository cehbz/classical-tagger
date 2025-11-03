package validation

import (
    "fmt"
    "strings"
    "unicode"

    "github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumBuilder provides a fluent API for building domain.Album instances.
type AlbumBuilder struct {
	album *domain.Album
}

// TrackBuilder provides a fluent API for building domain.Track instances.
type TrackBuilder struct {
	track   *domain.Track
	builder *AlbumBuilder // reference to parent builder for Build() to return
}

// NewAlbum creates a new AlbumBuilder with sensible defaults.
// Defaults: Title="Album", OriginalYear=1963, single track (Disc=1, Track=1, Title="Symphony")
// with Composer="Beethoven" and Ensemble="Orchestra".
func NewAlbum() *AlbumBuilder {
	return &AlbumBuilder{
		album: &domain.Album{
			Title:        "Album",
			OriginalYear: 0,
			Tracks: []*domain.Track{
				{
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

// WithTitle sets the album title.
func (b *AlbumBuilder) WithTitle(title string) *AlbumBuilder {
	b.album.Title = title
	return b
}

// WithOriginalYear sets the album's original year.
func (b *AlbumBuilder) WithOriginalYear(year int) *AlbumBuilder {
	b.album.OriginalYear = year
	return b
}

// WithComposer adds a composer to the default track (first track).
// If no tracks exist, creates a default track first.
func (b *AlbumBuilder) WithComposer(name string) *AlbumBuilder {
	b.ensureDefaultTrack()
	if len(b.album.Tracks) > 0 {
		b.album.Tracks[0].Artists = append(b.album.Tracks[0].Artists, domain.Artist{
			Name: name,
			Role: domain.RoleComposer,
		})
	}
	return b
}

// WithComposers adds multiple composers to the default track (variadic).
func (b *AlbumBuilder) WithComposers(names ...string) *AlbumBuilder {
	b.ensureDefaultTrack()
	if len(b.album.Tracks) > 0 {
		for _, name := range names {
			b.album.Tracks[0].Artists = append(b.album.Tracks[0].Artists, domain.Artist{
				Name: name,
				Role: domain.RoleComposer,
			})
		}
	}
	return b
}

// WithArtist adds a specific artist to the default track.
func (b *AlbumBuilder) WithArtist(name string, role domain.Role) *AlbumBuilder {
	b.ensureDefaultTrack()
	if len(b.album.Tracks) > 0 {
		b.album.Tracks[0].Artists = append(b.album.Tracks[0].Artists, domain.Artist{
			Name: name,
			Role: role,
		})
	}
	return b
}

// WithArtists adds multiple artists to the default track (variadic).
func (b *AlbumBuilder) WithArtists(artists ...domain.Artist) *AlbumBuilder {
	b.ensureDefaultTrack()
	if len(b.album.Tracks) > 0 {
		b.album.Tracks[0].Artists = append(b.album.Tracks[0].Artists, artists...)
	}
	return b
}

// WithEdition sets the album edition.
func (b *AlbumBuilder) WithEdition(label, catalogNumber string, year int) *AlbumBuilder {
	b.album.Edition = &domain.Edition{
		Label:         label,
		CatalogNumber: catalogNumber,
		Year:          year,
	}
	return b
}

// WithoutEdition explicitly removes the edition.
func (b *AlbumBuilder) WithoutEdition() *AlbumBuilder {
	b.album.Edition = nil
	return b
}

// AddTrack returns a TrackBuilder for adding a new track to the album.
func (b *AlbumBuilder) AddTrack() *TrackBuilder {
	return &TrackBuilder{
		track: &domain.Track{
			Disc:  1,
			Track: len(b.album.Tracks) + 1,
			Artists: []domain.Artist{
				{Name: "Beethoven", Role: domain.RoleComposer},
				{Name: "Orchestra", Role: domain.RoleEnsemble},
			},
		},
		builder: b,
	}
}

// AddTracks adds multiple tracks to the album (variadic).
func (b *AlbumBuilder) AddTracks(tracks ...*domain.Track) *AlbumBuilder {
	b.album.Tracks = append(b.album.Tracks, tracks...)
	return b
}

// WithTracks replaces all tracks in the album.
func (b *AlbumBuilder) WithTracks(tracks []*domain.Track) *AlbumBuilder {
	b.album.Tracks = tracks
	return b
}

// ClearTracks removes all tracks from the album.
func (b *AlbumBuilder) ClearTracks() *AlbumBuilder {
	b.album.Tracks = nil
	return b
}

// Build returns the constructed album.
func (b *AlbumBuilder) Build() *domain.Album {
	return b.album
}

// ensureDefaultTrack ensures there's at least one track with defaults.
func (b *AlbumBuilder) ensureDefaultTrack() {
	if len(b.album.Tracks) == 0 {
		b.album.Tracks = []*domain.Track{
			{
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
	tb.track.Name = filename
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

// Build adds the track to the album and returns the album builder.
func (tb *TrackBuilder) Build() *AlbumBuilder {
	tb.builder.album.Tracks = append(tb.builder.album.Tracks, tb.track)
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

// buildAlbumWithTitlesAndFilenames creates album with specific title/filename pairs
func buildAlbumWithTitlesAndFilenames(titleFiles []titleFile) *domain.Album {
	builder := NewAlbum().
		WithTitle("Test Album").
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

// Helper function to build an album with specific filenames
func buildAlbumWithFilenames(filenames ...string) *domain.Album {
	builder := NewAlbum().
		WithTitle("Beethoven Symphonies").
		ClearTracks()

	for i := range filenames {
		builder.AddTrack().
			WithTrack(i+1).
			WithTitle("Symphony No. 5").
			WithFilename(filenames[i]).
			ClearArtists().
			WithArtists(
				domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer},
				domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},
				domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			).
			Build()
	}

	return builder.Build()
}

// buildSingleComposerAlbumWithFilenames creates album with one composer
func buildSingleComposerAlbumWithFilenames(composerName string, filenames ...string) *domain.Album {
	builder := NewAlbum().ClearTracks()

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

// buildMultiComposerAlbumWithFilenames creates album with multiple composers
func buildMultiComposerAlbumWithFilenames(composerTracks ...[]composerTrack) *domain.Album {
	builder := NewAlbum().
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

// buildAlbumWithTrackFilenames creates an album with specific track/filename pairs
func buildAlbumWithTrackFilenames(trackFiles ...trackFile) *domain.Album {
	builder := NewAlbum().
		WithTitle("Test Album").
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

// buildMultiDiscAlbumWithFilenames creates multi-disc album
func buildMultiDiscAlbumWithFilenames(disc1, disc2 []trackFile) *domain.Album {
	builder := NewAlbum().
		WithTitle("Multi-Disc Album").
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

// buildAlbumWithDiscTracks creates an album with specific disc/track combinations
func buildAlbumWithDiscTracks(discTracks []discTrack) *domain.Album {
	builder := NewAlbum().
		WithTitle("Multi-Disc Album").
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

// buildAlbumWithArtistName creates album with specific artist name
func buildAlbumWithArtistName(artistName string) *domain.Album {
	return NewAlbum().
		WithComposer(artistName).
		Build()
}

// buildAlbumWithTrackTitle creates album with specific track title
func buildAlbumWithTrackTitle(title string) *domain.Album {
	return NewAlbum().
		ClearTracks().
		AddTrack().
		WithTitle(title).
		Build().
		Build()
}

// buildAlbumWithFilenamesAndDiscs creates album with specific filenames and disc numbers
func buildAlbumWithFilenamesAndDiscs(filenames []string, discs []int) *domain.Album {
	builder := NewAlbum().ClearTracks()

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

// buildAlbumWithCompleteEdition creates album with complete edition information
func buildAlbumWithCompleteEdition() *domain.Album {
	return NewAlbum().
		WithEdition("Deutsche Grammophon", "DG-479-0334", 1990).
		Build()
}

// buildAlbumWithConsistentSoloist creates album with same soloist on all tracks
func buildAlbumWithConsistentSoloist(soloistName string, trackCount int) *domain.Album {
	builder := NewAlbum().ClearTracks()

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

// buildAlbumWithGuestSoloist creates album where one soloist appears infrequently
func buildAlbumWithGuestSoloist(mainSoloist, guestSoloist string, trackCount int) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	main := domain.Artist{Name: mainSoloist, Role: domain.RoleSoloist}
	guest := domain.Artist{Name: guestSoloist, Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

	builder := NewAlbum().ClearTracks()

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

// buildAlbumWithGuestInTitle creates album with guest indicated in title
func buildAlbumWithGuestInTitle(mainSoloist, guestSoloist string, trackCount int) *domain.Album {
	composer := domain.Artist{Name: "Beethoven", Role: domain.RoleComposer}
	main := domain.Artist{Name: mainSoloist, Role: domain.RoleSoloist}
	guest := domain.Artist{Name: guestSoloist, Role: domain.RoleSoloist}
	ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

	builder := NewAlbum().ClearTracks()

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

// buildAlbumWithGoodCaps creates album with proper capitalization
func buildAlbumWithGoodCaps() *domain.Album {
	builder := NewAlbum().
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

// buildAlbumWithBadCaps creates album with poor capitalization
func buildAlbumWithBadCaps() *domain.Album {
	builder := NewAlbum().
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
