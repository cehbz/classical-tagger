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

	data := result.Data()

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
	if len(data.Tracks) == 0 {
		t.Error("No tracks extracted")
	}

	// Verify track structure
	for i, track := range data.Tracks {
		if track.Title == "" {
			t.Errorf("Track %d has no title", i+1)
		}
		if track.Composer == "" {
			t.Errorf("Track %d has no composer", i+1)
		}
		if track.Track != i+1 {
			t.Errorf("Track %d has wrong track number: got %d", i+1, track.Track)
		}
	}
}

func TestPrestoParser_ParseTitle(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		want    string
		wantErr bool
	}{
		{
			name: "title with og:title meta tag - should prefer this",
			html: `
			<html>
			<head>
				<title>RIAS Kammerchor: Christmas! - Harmonia Mundi: HMC902170 | Presto Music</title>
				<meta property="og:title" content="Noël! Christmas! Weihnachten!" />
			</head>
			</html>`,
			want:    "Noël! Christmas! Weihnachten!",
			wantErr: false,
		},
		{
			name: "title with h1 product block title",
			html: `
			<html>
			<head>
				<title>Something - Label | Presto Music</title>
			</head>
			<body>
				<h1 class="c-product-block__title">Actual Album Title</h1>
			</body>
			</html>`,
			want:    "Actual Album Title",
			wantErr: false,
		},
		{
			name:    "standard title - fallback when no og:title",
			html:    `<title>RIAS Kammerchor: Christmas! - Harmonia Mundi: HMC902170 - CD or download | Presto Music</title>`,
			want:    "RIAS Kammerchor: Christmas!",
			wantErr: false,
		},
		{
			name:    "title with special characters",
			html:    `<title>Artist: Album – Title | Presto Music</title>`,
			want:    "Artist: Album – Title",
			wantErr: false,
		},
		{
			name:    "no title tag",
			html:    `<html><body>No title</body></html>`,
			want:    "",
			wantErr: true,
		},
	}

	parser := NewPrestoParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseTitle(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrestoParser_ParseCatalogNumber(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		want    string
		wantErr bool
	}{
		{
			name:    "catalog in product metadata",
			html:    `<ul class="c-product-block__metadata"><li><strong>Catalogue number:</strong> HMC902170</li></ul>`,
			want:    "HMC902170",
			wantErr: false,
		},
		{
			name:    "no catalog number",
			html:    `<ul class="c-product-block__metadata"><li>Some other info</li></ul>`,
			want:    "",
			wantErr: true,
		},
	}

	parser := NewPrestoParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseCatalogNumber(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCatalogNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCatalogNumber() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrestoParser_ParseCatalogAndLabel(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		wantCatalog string
		wantLabel   string
		wantErr     bool
	}{
		{
			name: "catalog and label in metadata",
			html: `<ul class="c-product-block__metadata">
				<li><strong>Catalogue number:</strong> HMC902170</li>
				<li><strong>Label: </strong><a href="/labels/harmonia-mundi">Harmonia Mundi</a></li>
			</ul>`,
			wantCatalog: "HMC902170",
			wantLabel:   "Harmonia Mundi",
			wantErr:     false,
		},
		{
			name: "only catalog",
			html: `<ul class="c-product-block__metadata">
				<li><strong>Catalogue number:</strong> HMC902170</li>
			</ul>`,
			wantCatalog: "HMC902170",
			wantLabel:   "",
			wantErr:     false,
		},
		{
			name: "only label",
			html: `<ul class="c-product-block__metadata">
				<li><strong>Label: </strong><a href="/labels/harmonia-mundi">Harmonia Mundi</a></li>
			</ul>`,
			wantCatalog: "",
			wantLabel:   "Harmonia Mundi",
			wantErr:     false,
		},
		{
			name:        "no catalog or label",
			html:        `<ul class="c-product-block__metadata"><li>Other info</li></ul>`,
			wantCatalog: "",
			wantLabel:   "",
			wantErr:     true,
		},
	}

	parser := NewPrestoParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCatalog, gotLabel, err := parser.ParseCatalogAndLabel(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCatalogAndLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCatalog != tt.wantCatalog {
				t.Errorf("ParseCatalogAndLabel() catalog = %q, want %q", gotCatalog, tt.wantCatalog)
			}
			if gotLabel != tt.wantLabel {
				t.Errorf("ParseCatalogAndLabel() label = %q, want %q", gotLabel, tt.wantLabel)
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
	if tracks[0].Composer != "Mendelssohn" {
		t.Errorf("Track 1 composer = %q, want %q", tracks[0].Composer, "Mendelssohn")
	}
	if tracks[0].Title != "Frohlocket, ihr Völker auf Erden, Op.79 No.1" {
		t.Errorf("Track 1 title = %q, want %q", tracks[0].Title, "Frohlocket, ihr Völker auf Erden, Op.79 No.1")
	}

	// Check second track
	if tracks[1].Composer != "Johann Eccard" {
		t.Errorf("Track 2 composer = %q, want %q", tracks[1].Composer, "Johann Eccard")
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
			t.Logf("  %d. %s (composer: %s)", i+1, track.Title, track.Composer)
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
		if track.Composer != "Poulenc" {
			t.Errorf("Track %d composer = %q, want %q", i+1, track.Composer, "Poulenc")
		}
	}

	// Verify last track is standalone work (not part of Poulenc cycle)
	if len(tracks) >= 5 {
		lastTrack := tracks[4]
		if !strings.Contains(lastTrack.Title, "Stille Nacht") {
			t.Errorf("Track 5 should be 'Stille Nacht', got %q", lastTrack.Title)
		}
		if lastTrack.Composer != "Mandyczewski" {
			t.Errorf("Track 5 composer = %q, want %q", lastTrack.Composer, "Mandyczewski")
		}
	}
}