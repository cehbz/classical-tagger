package domain_test

import (
	"fmt"
	"github.com/cehbz/classical-tagger/internal/domain"
)

// ExampleAlbum demonstrates how to create and validate a complete classical music album.
func ExampleAlbum() {
	// Create an album
	album, err := domain.NewAlbum("Noël ! Weihnachten ! Christmas!", 2013)
	if err != nil {
		panic(err)
	}
	
	// Add edition information
	edition, _ := domain.NewEdition("harmonia mundi", 2013)
	edition = edition.WithCatalogNumber("HMC902170")
	album = album.WithEdition(edition)
	
	// Create artists
	mendelssohn, _ := domain.NewArtist("Felix Mendelssohn Bartholdy", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("RIAS Kammerchor Berlin", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Hans-Christoph Rademann", domain.RoleConductor)
	
	// Create and add tracks
	track1, _ := domain.NewTrack(
		1, // disc
		1, // track number
		"Frohlocket, ihr Völker auf Erden, Op. 79/1",
		[]domain.Artist{mendelssohn, ensemble, conductor},
	)
	track1 = track1.WithName("01 Frohlocket, ihr Völker auf Erden, Op. 79-1.flac")
	album.AddTrack(track1)
	
	brahms, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
	track2, _ := domain.NewTrack(
		1,
		2,
		"O Heiland, reiß die Himmel auf, Op. 74/2",
		[]domain.Artist{brahms, ensemble, conductor},
	)
	track2 = track2.WithName("02 O Heiland, reiß die Himmel auf, Op. 74-2.flac")
	album.AddTrack(track2)
	
	// Validate the album
	issues := album.Validate()
	
	if len(issues) == 0 {
		fmt.Println("Album is valid!")
	} else {
		fmt.Printf("Found %d validation issues:\n", len(issues))
		for _, issue := range issues {
			fmt.Println(issue)
		}
	}
	
	// Output:
	// Album is valid!
}

// ExampleAlbum_withErrors demonstrates validation errors.
func ExampleAlbum_withErrors() {
	album, _ := domain.NewAlbum("Test Album", 2013)
	
	// Create a track with composer name in title (ERROR)
	bach, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	badTrack, _ := domain.NewTrack(1, 1, "Bach: Goldberg Variations", []domain.Artist{bach})
	album.AddTrack(badTrack)
	
	// Validate
	issues := album.Validate()
	
	for _, issue := range issues {
		fmt.Printf("[%s] %s\n", issue.Level(), issue.Message())
	}
	
	// Output:
	// [WARNING] Edition information (label, catalog number) is strongly recommended
	// [ERROR] Composer name 'Johann Sebastian Bach' must not appear in track title tag
}

// ExampleTrack_multipleComposers demonstrates the multiple composer error.
func ExampleTrack_multipleComposers() {
	composer1, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	composer2, _ := domain.NewArtist("Carl Philipp Emanuel Bach", domain.RoleComposer)
	
	_, err := domain.NewTrack(
		1,
		1,
		"Some Work",
		[]domain.Artist{composer1, composer2},
	)
	
	if err != nil {
		fmt.Println(err)
	}
	
	// Output:
	// multiple composers not supported (found 2)
}

// ExampleTrack_arrangement demonstrates parsing arranger from title.
func ExampleTrack_arrangement() {
	// In the future, we would parse this automatically
	title := "Goldberg Variations (arr. by Sitkovetsky)"
	
	// For now, we handle it manually:
	composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	arranger, _ := domain.NewArtist("Dmitry Sitkovetsky", domain.RoleArranger)
	
	track, _ := domain.NewTrack(
		1,
		1,
		title,
		[]domain.Artist{composer, arranger},
	)
	
	fmt.Printf("Track: %s\n", track.Title())
	fmt.Printf("Composer: %s\n", track.Composer().Name())
	
	// Find arranger
	for _, artist := range track.Artists() {
		if artist.Role() == domain.RoleArranger {
			fmt.Printf("Arranger: %s\n", artist.Name())
		}
	}
	
	// Output:
	// Track: Goldberg Variations (arr. by Sitkovetsky)
	// Composer: Johann Sebastian Bach
	// Arranger: Dmitry Sitkovetsky
}
