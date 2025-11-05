package scraping

import (
	"os"
	"strings"
	"testing"
)

func TestPrestoParser_Parse(t *testing.T) {
	// Read test HTML file
	html, err := os.ReadFile("testdata/presto_christmas.html")
	if err != nil {
		t.Skipf("Test HTML file not available: %v", err)
	}

	parser := NewPrestoParser()
	result, err := parser.Parse(string(html))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil result")
	}

	data := result.Torrent

	// Test title extraction
	if data.Title == "" || data.Title == MissingTitle {
		t.Error("Title not extracted")
	}

	// Test year extraction
	if data.OriginalYear == 0 || data.OriginalYear == MissingYear {
		t.Error("Year not extracted")
	}

	// Test catalog number extraction
	if data.Edition == nil || data.Edition.CatalogNumber == "" {
		t.Error("Catalog number not extracted")
	}

	// Test tracks extraction
	tracks := data.Tracks()
	if len(tracks) == 0 {
		t.Error("No tracks extracted")
	}

	// Verify track structure
	for i, track := range tracks {
		if track.Title == "" {
			t.Errorf("Track %d has no title", i+1)
		}
		if len(track.Composers()) == 0 {
			t.Errorf("Track %d has no composer", i+1)
		}
		if track.Track != i+1 {
			t.Errorf("Track %d has wrong track number: got %d", i+1, track.Track)
		}
	}
}

func TestPrestoParser_ParseTitle(t *testing.T) {
	tests := []struct {
		Name    string
		HTML    string
		Want    string
		WantErr bool
	}{
		{
			Name: "title with og:title meta tag - should prefer this",
			HTML: `
			<html>
			<head>
				<title>RIAS Kammerchor: Christmas! - Harmonia Mundi: HMC902170 | Presto Music</title>
				<meta property="og:title" content="Noël! Christmas! Weihnachten!" />
			</head>
			</html>`,
			Want:    "Noël! Christmas! Weihnachten!",
			WantErr: false,
		},
		{
			Name: "title with h1 product block title",
			HTML: `
			<html>
			<head>
				<title>Something - Label | Presto Music</title>
			</head>
			<body>
				<h1 class="c-product-block__title">Actual Album Title</h1>
			</body>
			</html>`,
			Want:    "Actual Album Title",
			WantErr: false,
		},
		{
			Name:    "standard title - fallback when no og:title",
			HTML:    `<title>RIAS Kammerchor: Christmas! - Harmonia Mundi: HMC902170 - CD or download | Presto Music</title>`,
			Want:    "RIAS Kammerchor: Christmas!",
			WantErr: false,
		},
		{
			Name:    "title with special characters",
			HTML:    `<title>Artist: Album – Title | Presto Music</title>`,
			Want:    "Artist: Album – Title",
			WantErr: false,
		},
		{
			Name:    "no title tag",
			HTML:    `<html><body>No title</body></html>`,
			Want:    "",
			WantErr: true,
		},
	}

	parser := NewPrestoParser()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := parser.ParseTitle(tt.HTML)
			if (err != nil) != tt.WantErr {
				t.Errorf("ParseTitle() error = %v, wantErr %v", err, tt.WantErr)
				return
			}
			if got != tt.Want {
				t.Errorf("ParseTitle() = %q, want %q", got, tt.Want)
			}
		})
	}
}

func TestPrestoParser_ParseCatalogNumber(t *testing.T) {
	tests := []struct {
		Name    string
		HTML    string
		Want    string
		WantErr bool
	}{
		{
			Name:    "catalog in product metadata",
			HTML:    `<ul class="c-product-block__metadata"><li><strong>Catalogue number:</strong> HMC902170</li></ul>`,
			Want:    "HMC902170",
			WantErr: false,
		},
		{
			Name:    "no catalog number",
			HTML:    `<ul class="c-product-block__metadata"><li>Some other info</li></ul>`,
			Want:    "",
			WantErr: true,
		},
	}

	parser := NewPrestoParser()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := parser.ParseCatalogNumber(tt.HTML)
			if (err != nil) != tt.WantErr {
				t.Errorf("ParseCatalogNumber() error = %v, wantErr %v", err, tt.WantErr)
				return
			}
			if got != tt.Want {
				t.Errorf("ParseCatalogNumber() = %q, want %q", got, tt.Want)
			}
		})
	}
}

func TestPrestoParser_ParseCatalogAndLabel(t *testing.T) {
	tests := []struct {
		Name        string
		HTML        string
		WantCatalog string
		WantLabel   string
		WantErr     bool
	}{
		{
			Name: "catalog and label in metadata",
			HTML: `<ul class="c-product-block__metadata">
				<li><strong>Catalogue number:</strong> HMC902170</li>
				<li><strong>Label: </strong><a href="/labels/harmonia-mundi">Harmonia Mundi</a></li>
			</ul>`,
			WantCatalog: "HMC902170",
			WantLabel:   "Harmonia Mundi",
			WantErr:     false,
		},
		{
			Name: "only catalog",
			HTML: `<ul class="c-product-block__metadata">
				<li><strong>Catalogue number:</strong> HMC902170</li>
			</ul>`,
			WantCatalog: "HMC902170",
			WantLabel:   "",
			WantErr:     false,
		},
		{
			Name: "only label",
			HTML: `<ul class="c-product-block__metadata">
				<li><strong>Label: </strong><a href="/labels/harmonia-mundi">Harmonia Mundi</a></li>
			</ul>`,
			WantCatalog: "",
			WantLabel:   "Harmonia Mundi",
			WantErr:     false,
		},
		{
			Name:        "no catalog or label",
			HTML:        `<ul class="c-product-block__metadata"><li>Other info</li></ul>`,
			WantCatalog: "",
			WantLabel:   "",
			WantErr:     true,
		},
	}

	parser := NewPrestoParser()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gotCatalog, gotLabel, err := parser.ParseCatalogAndLabel(tt.HTML)
			if (err != nil) != tt.WantErr {
				t.Errorf("ParseCatalogAndLabel() error = %v, wantErr %v", err, tt.WantErr)
				return
			}
			if gotCatalog != tt.WantCatalog {
				t.Errorf("ParseCatalogAndLabel() catalog = %q, want %q", gotCatalog, tt.WantCatalog)
			}
			if gotLabel != tt.WantLabel {
				t.Errorf("ParseCatalogAndLabel() label = %q, want %q", gotLabel, tt.WantLabel)
			}
		})
	}
}

func TestPrestoParser_ParseTracks(t *testing.T) {
	html := `
	<div class="c-tracklist__work">
		<div class="c-track__title">
			<a href="/composer">Mendelssohn</a>:
			<a href="/works">Frohlocket, ihr Völker auf Erden, Op.79 No.1</a>
		</div>
	</div>
	<div class="c-tracklist__work">
		<div class="c-track__title">
			Johann Eccard:
			Ich lag in tiefster Todesnacht
		</div>
	</div>
	`

	parser := NewPrestoParser()
	tracks, err := parser.ParseTracks(html)

	if err != nil {
		t.Fatalf("ParseTracks() error = %v", err)
	}

	if len(tracks) != 2 {
		t.Fatalf("ParseTracks() got %d tracks, want 2", len(tracks))
	}

	// Check first track
	composers := tracks[0].Composers()
	if len(composers) != 1 {
		t.Errorf("Track 1 has %d composers, want 1", len(composers))
	}
	composer := composers[0]
	if composer.Name != "Mendelssohn" {
		t.Errorf("Track 1 composer = %q, want %q", composer.Name, "Mendelssohn")
	}
	if tracks[0].Title != "Frohlocket, ihr Völker auf Erden, Op.79 No.1" {
		t.Errorf("Track 1 title = %q, want %q", tracks[0].Title, "Frohlocket, ihr Völker auf Erden, Op.79 No.1")
	}

	// Check second track
	composers = tracks[1].Composers()
	if len(composers) != 1 {
		t.Errorf("Track 2 has %d composers, want 1", len(composers))
	}
	composer = composers[0]
	if composer.Name != "Johann Eccard" {
		t.Errorf("Track 2 composer = %q, want %q", composer.Name, "Johann Eccard")
	}
	if tracks[1].Title != "Ich lag in tiefster Todesnacht" {
		t.Errorf("Track 2 title = %q, want %q", tracks[1].Title, "Ich lag in tiefster Todesnacht")
	}
}

// TestPrestoParser_ParseTracks_FlattenHierarchy tests that hierarchical work structures
// (parent work + movements) are flattened to match physical disc organization.
func TestPrestoParser_ParseTracks_FlattenHierarchy(t *testing.T) {
	html := `
	<html>
	<body>
		<div class="c-tracklist">
			<!-- Hierarchical work with movements -->
			<div class="c-tracklist__work">
				<div class="c-track__title">
					<a href="/composer/poulenc">Poulenc</a>:
					<a href="/works/quatre-motets">Quatre motets pour le temps de Noël</a>
				</div>
				
				<!-- Movement 1 -->
				<div class="c-track--track">
					<div class="c-track__title">I. O magnum mysterium</div>
				</div>
				
				<!-- Movement 2 -->
				<div class="c-track--track">
					<div class="c-track__title">II. Quem vidistis pastores dicite</div>
				</div>
				
				<!-- Movement 3 -->
				<div class="c-track--track">
					<div class="c-track__title">III. Videntes stellam</div>
				</div>
				
				<!-- Movement 4 -->
				<div class="c-track--track">
					<div class="c-track__title">IV. Hodie Christus natus est</div>
				</div>
			</div>
			
			<!-- Next standalone work -->
			<div class="c-tracklist__work">
				<div class="c-track__title">
					<a href="/composer/mandyczewski">Mandyczewski</a>:
					Stille Nacht, heilige Nacht
				</div>
			</div>
		</div>
	</body>
	</html>
	`

	parser := NewPrestoParser()
	tracks, err := parser.ParseTracks(html)

	if err != nil {
		t.Fatalf("ParseTracks() error = %v", err)
	}

	// Should have 5 tracks total (4 Poulenc movements + 1 Mandyczewski)
	// NOT 6 tracks (parent + 4 movements + 1)
	if len(tracks) != 4 {
		t.Errorf("Got %d tracks, want 4 (hierarchy should be flattened)", len(tracks))
		t.Logf("Tracks extracted:")
		for i, track := range tracks {
			for _, composer := range track.Composers() {
				t.Logf("  %d. %s (composer: %s)", i+1, track.Title, composer.Name)
			}
		}
	}

	// Check that Poulenc tracks have cycle name prepended
	expectedPoulencTitles := []string{
		"Quatre motets pour le temps de Noël: I. O magnum mysterium",
		"Quatre motets pour le temps de Noël: II. Quem vidistis pastores dicite",
		"Quatre motets pour le temps de Noël: III. Videntes stellam",
		"Quatre motets pour le temps de Noël: IV. Hodie Christus natus est",
	}

	for i := 0; i < 4 && i < len(tracks); i++ {
		track := tracks[i]
		expected := expectedPoulencTitles[i]

		if track.Title != expected {
			t.Errorf("Track %d title = %q, want %q", i+1, track.Title, expected)
		}

		// Verify composer is preserved from parent
		composers := track.Composers()
		if len(composers) != 1 {
			t.Errorf("Track %d has %d composers, want 1", i+1, len(composers))
		}
		composer := composers[0]
		if composer.Name != "Poulenc" {
			t.Errorf("Track %d composer = %q, want %q", i+1, composer.Name, "Poulenc")
		}
	}

	// Verify last track is standalone work (not part of Poulenc cycle)
	if len(tracks) >= 5 {
		lastTrack := tracks[4]
		if !strings.Contains(lastTrack.Title, "Stille Nacht") {
			t.Errorf("Track 5 should be 'Stille Nacht', got %q", lastTrack.Title)
		}
		composers := lastTrack.Composers()
		if len(composers) != 1 {
			t.Errorf("Track 5 has %d composers, want 1", len(composers))
		}
		composer := composers[0]
		if composer.Name != "Mandyczewski" {
			t.Errorf("Track 5 composer = %q, want %q", composer.Name, "Mandyczewski")
		}
	}
}
