package scraping

import (
	"os"
	"testing"
)

func TestPrestoClassicalParser_ParseTitle(t *testing.T) {
	html := `
		<html>
		<head><title>Test</title></head>
		<body>
			<h1>Nature at Play: Bach's Cello Suite No. 1 (Live from the Great Smoky Mountains) (Digital Download)</h1>
		</body>
		</html>
	`

	parser := NewPrestoClassicalParser()
	title, err := parser.ParseTitle(html)

	if err != nil {
		t.Fatalf("ParseTitle() error = %v", err)
	}

	expected := "Nature at Play: Bach's Cello Suite No. 1 (Live from the Great Smoky Mountains)"
	if title != expected {
		t.Errorf("ParseTitle() = %q, want %q", title, expected)
	}
}

func TestPrestoClassicalParser_ParseTitle_Missing(t *testing.T) {
	html := `<html><body></body></html>`

	parser := NewPrestoClassicalParser()
	_, err := parser.ParseTitle(html)

	if err == nil {
		t.Error("ParseTitle() expected error for missing title, got nil")
	}
}

func TestPrestoClassicalParser_ParseYear(t *testing.T) {
	tests := []struct {
		name string
		html string
		want int
	}{
		{
			name: "year in product info",
			html: `<div class="c-product__info">Recorded in 2020</div>`,
			want: 2020,
		},
		{
			name: "year in product details",
			html: `<div class="c-product-details">Released: 1995</div>`,
			want: 1995,
		},
		{
			name: "no year found",
			html: `<div>No year here</div>`,
			want: 0,
		},
	}

	parser := NewPrestoClassicalParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseYear(tt.html)
			if tt.want == 0 {
				if err == nil {
					t.Error("ParseYear() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("ParseYear() error = %v", err)
				}
				if got != tt.want {
					t.Errorf("ParseYear() = %d, want %d", got, tt.want)
				}
			}
		})
	}
}

func TestPrestoClassicalParser_ParseTracks(t *testing.T) {
	html := `
		<html>
		<body>
			<div class="c-tracklist">
				<div class="c-tracklist__work">
					<div class="c-track c-track--work">
						<a href="/composer/bach">Bach, J S</a>
						<a class="c-track__title" href="/work/cello-suite-1">Cello Suite No. 1 in G major, BWV1007</a>
					</div>
					<div class="c-track c-track--track">
						<span class="c-track__title">I. Prélude</span>
						<span class="c-track__duration">Track length2:33</span>
						<span class="c-track__performers">Yo-Yo Ma (cello)</span>
					</div>
					<div class="c-track c-track--track">
						<span class="c-track__title">II. Allemande</span>
						<span class="c-track__duration">Track length2:02</span>
						<span class="c-track__performers">Yo-Yo Ma (cello)</span>
					</div>
				</div>
			</div>
		</body>
		</html>
	`

	parser := NewPrestoClassicalParser()
	tracks, errors := parser.ParseTracks(html)

	if len(errors) > 0 {
		for _, err := range errors {
			t.Logf("Warning: %v", err)
		}
	}

	if len(tracks) != 2 {
		t.Fatalf("ParseTracks() got %d tracks, want 2", len(tracks))
	}

	// Check first track
	if tracks[0].Track != 1 {
		t.Errorf("Track 1 number = %d, want 1", tracks[0].Track)
	}
	if tracks[0].Title != "I. Prélude" {
		t.Errorf("Track 1 title = %q, want %q", tracks[0].Title, "I. Prélude")
	}
	if tracks[0].Composer != "Bach, J S" {
		t.Errorf("Track 1 composer = %q, want %q", tracks[0].Composer, "Bach, J S")
	}
	if len(tracks[0].Artists) != 1 {
		t.Errorf("Track 1 artists = %d, want 1", len(tracks[0].Artists))
	} else {
		if tracks[0].Artists[0].Name != "Yo-Yo Ma (cello)" {
			t.Errorf("Track 1 artist name = %q, want %q", tracks[0].Artists[0].Name, "Yo-Yo Ma (cello)")
		}
	}

	// Check second track
	if tracks[1].Track != 2 {
		t.Errorf("Track 2 number = %d, want 2", tracks[1].Track)
	}
	if tracks[1].Title != "II. Allemande" {
		t.Errorf("Track 2 title = %q, want %q", tracks[1].Title, "II. Allemande")
	}
}

func TestPrestoClassicalParser_ParseEdition(t *testing.T) {
	html := `
		<html>
		<body>
			<script type="application/ld+json">
			{
				"@type": "Product",
				"mpn": "G010005113879S",
				"brand": {"name": "Sony Classical"}
			}
			</script>
		</body>
		</html>
	`

	parser := NewPrestoClassicalParser()
	edition, err := parser.ParseEdition(html)

	if err != nil {
		t.Fatalf("ParseEdition() error = %v", err)
	}

	if edition.CatalogNumber != "G010005113879S" {
		t.Errorf("CatalogNumber = %q, want %q", edition.CatalogNumber, "G010005113879S")
	}

	if edition.Label != "Sony Classical" {
		t.Errorf("Label = %q, want %q", edition.Label, "Sony Classical")
	}
}

func TestPrestoClassicalParser_Parse_RealHTML(t *testing.T) {
	// Load real HTML sample if available
	htmlBytes, err := os.ReadFile("/mnt/user-data/uploads/prestoclassical_example.html")
	if err != nil {
		t.Skip("Real HTML sample not available")
	}

	html := string(htmlBytes)
	parser := NewPrestoClassicalParser()
	result, err := parser.Parse(html)

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	data := result.Data()

	// Validate basic fields
	if data.Title == "" || data.Title == MissingTitle {
		t.Error("Title not extracted")
	}
	t.Logf("Title: %s", data.Title)

	if len(data.Tracks) == 0 {
		t.Error("No tracks extracted")
	}
	t.Logf("Tracks: %d", len(data.Tracks))

	// Validate track structure
	for i, track := range data.Tracks {
		if track.Track != i+1 {
			t.Errorf("Track %d has wrong number: %d", i, track.Track)
		}
		if track.Title == "" {
			t.Errorf("Track %d has no title", i)
		}
		if track.Composer == "" {
			t.Logf("Warning: Track %d has no composer", i)
		}
	}

	// Check for errors
	if result.HasRequiredErrors() {
		t.Error("Parse() has required field errors:")
		for _, e := range result.Errors() {
			if e.Required() {
				t.Errorf("  - %s: %s", e.Field(), e.Message())
			}
		}
	}

	// Log warnings
	for _, w := range result.Warnings() {
		t.Logf("Warning: %s", w)
	}

	// Try to convert to domain
	_, err = data.ToAlbum()
	if err != nil {
		t.Logf("Warning: Domain conversion failed: %v", err)
	}
}

func TestPrestoClassicalParser_InferArtistRole(t *testing.T) {
	tests := []struct {
		name     string
		artist   string
		context  string
		wantRole string
	}{
		{
			name:     "orchestra",
			artist:   "Berlin Philharmonic Orchestra",
			context:  "",
			wantRole: "ensemble",
		},
		{
			name:     "conductor",
			artist:   "Herbert von Karajan, conductor",
			context:  "",
			wantRole: "conductor",
		},
		{
			name:     "soloist with piano",
			artist:   "Martha Argerich",
			context:  "Piano Concerto No. 1",
			wantRole: "soloist",
		},
		{
			name:     "unknown role",
			artist:   "John Doe",
			context:  "Some piece",
			wantRole: "performer",
		},
		{
			name:     "quartet",
			artist:   "Emerson String Quartet",
			context:  "",
			wantRole: "ensemble",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferArtistRole(tt.artist, tt.context)
			if got != tt.wantRole {
				t.Errorf("inferArtistRole(%q, %q) = %q, want %q", 
					tt.artist, tt.context, got, tt.wantRole)
			}
		})
	}
}

func TestPrestoClassicalParser_Parse_EmptyHTML(t *testing.T) {
	parser := NewPrestoClassicalParser()
	result, err := parser.Parse("")

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !result.HasRequiredErrors() {
		t.Error("Parse() should have required errors for empty HTML")
	}
}

func TestPrestoClassicalParser_ParseTracks_NoTracklist(t *testing.T) {
	html := `<html><body><div>No tracklist here</div></body></html>`

	parser := NewPrestoClassicalParser()
	tracks, errors := parser.ParseTracks(html)

	if len(tracks) > 0 {
		t.Errorf("ParseTracks() got %d tracks, want 0", len(tracks))
	}

	if len(errors) == 0 {
		t.Error("ParseTracks() expected errors, got none")
	}
}