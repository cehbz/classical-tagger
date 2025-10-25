package scraping

import (
	"os"
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
			name: "standard title",
			html: `<title>RIAS Kammerchor: Christmas! - Harmonia Mundi: HMC902170 - CD or download | Presto Music</title>`,
			want: "RIAS Kammerchor: Christmas!",
			wantErr: false,
		},
		{
			name: "title with special characters",
			html: `<title>Artist: Album – Title | Presto Music</title>`,
			want: "Artist: Album – Title",
			wantErr: false,
		},
		{
			name: "no title tag",
			html: `<html><body>No title</body></html>`,
			want: "",
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
			name: "catalog in product metadata",
			html: `<ul class="c-product-block__metadata"><li><strong>Catalogue number:</strong> HMC902170</li></ul>`,
			want: "HMC902170",
			wantErr: false,
		},
		{
			name: "no catalog number",
			html: `<ul class="c-product-block__metadata"><li>Some other info</li></ul>`,
			want: "",
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