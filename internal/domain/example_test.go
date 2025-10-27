package domain_test

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ExampleAlbum demonstrates how to create and validate a complete classical music album.
func ExampleAlbum() {
	// Create an album
	album := domain.Album{
		Title: "Noël ! Weihnachten ! Christmas!",
		OriginalYear: 2013,
		Edition: &domain.Edition{
			Label: "test label",
			Year: 2013,
			CatalogNumber: "HMC902170",
		},
		Tracks: []*domain.Track{
			&domain.Track{
				Disc: 1,
				Track: 1,
				Title: "Frohlocket, ihr Völker auf Erden, Op. 79/1",
				Artists: []domain.Artist{
					domain.Artist{Name: "Felix Mendelssohn Bartholdy", Role: domain.RoleComposer},
					domain.Artist{Name: "RIAS Kammerchor Berlin", Role: domain.RoleEnsemble},
					domain.Artist{Name: "Hans-Christoph Rademann", Role: domain.RoleConductor},
				},
				Name:  "01 Frohlocket, ihr Völker auf Erden, Op. 79-1.flac",
			},
			&domain.Track{
				Disc: 1,
				Track: 2,
				Title: "O Heiland, reiß die Himmel auf, Op. 74/2",
				Artists: []domain.Artist{
					domain.Artist{Name: "Johannes Brahms", Role: domain.RoleComposer},
					domain.Artist{Name: "RIAS Kammerchor Berlin", Role: domain.RoleEnsemble},
					domain.Artist{Name: "Hans-Christoph Rademann", Role: domain.RoleConductor},
				},
				Name: "02 O Heiland, reiß die Himmel auf, Op. 74-2.flac",
			},
		},
	}

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
	album := domain.Album{
		Title: "Test Album", OriginalYear: 2013, Tracks: []*domain.Track{
			&domain.Track{
				Disc: 1,
				Track: 1,
				Title: "Bach: Goldberg Variations",
				Artists: []domain.Artist{domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer}},
			},
		},
	}

	// Validate
	issues := album.Validate()

	for _, issue := range issues {
		fmt.Printf("[%s] %s\n", issue.Level, issue.Message)
	}

	// Output:
	// [WARNING] Edition information (label, catalog number) is strongly recommended
	// [ERROR] Composer name 'Johann Sebastian Bach' must not appear in track title tag
}

// ExampleTrack_multipleComposers demonstrates the multiple composer error.
func ExampleTrack_multipleComposers() {
	tr := domain.Track{
		Disc: 1,
		Track: 1,
		Title: "Some Work",
		Artists: []domain.Artist{
			domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer},
			domain.Artist{Name: "Carl Philipp Emanuel Bach", Role: domain.RoleComposer},
		},
		Name: "01 Some Work.flac",
	}
	fmt.Println(tr.Title)
	// Output:
	// Some Work
}

// ExampleTrack_arrangement demonstrates parsing arranger from title.
func ExampleTrack_arrangement() {
	// In the future, we would parse this automatically
	track := domain.Track{
		Disc: 1,
		Track: 1,
		Title: "Goldberg Variations (arr. by Sitkovetsky)",
		Artists: []domain.Artist{
			domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer},
			domain.Artist{Name: "Dmitry Sitkovetsky", Role: domain.RoleArranger},
		},
		Name: "01 Goldberg Variations.flac",
	}

	fmt.Printf("Track: %s\n", track.Title)
	for _, composer := range track.Composers() {
		fmt.Printf("Composer: %s\n", composer.Name)
	}

	// Find arranger
	for _, artist := range track.Artists {
		if artist.Role == domain.RoleArranger {
			fmt.Printf("Arranger: %s\n", artist.Name)
		}
	}

	// Output:
	// Track: Goldberg Variations (arr. by Sitkovetsky)
	// Composer: Johann Sebastian Bach
	// Arranger: Dmitry Sitkovetsky
}
