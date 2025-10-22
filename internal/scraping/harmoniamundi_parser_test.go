package scraping

import (
	"os"
	"testing"
)

// TestHarmoniaMundiParser_ParseTitle tests album title extraction
func TestHarmoniaMundiParser_ParseTitle(t *testing.T) {
	html := `<title>Noël ! Weihnachten ! Christmas! | harmonia mundi</title>`

	parser := NewHarmoniaMundiParser()
	title, err := parser.ParseTitle(html)

	if err != nil {
		t.Fatalf("ParseTitle() error = %v", err)
	}

	expected := "Noël ! Weihnachten ! Christmas!"
	if title != expected {
		t.Errorf("ParseTitle() = %q, want %q", title, expected)
	}
}

// TestHarmoniaMundiParser_ParseYear tests year extraction from JSON-LD
func TestHarmoniaMundiParser_ParseYear(t *testing.T) {
	html := `"datePublished": "October 2013"`

	parser := NewHarmoniaMundiParser()
	year, err := parser.ParseYear(html)

	if err != nil {
		t.Fatalf("ParseYear() error = %v", err)
	}

	if year != 2013 {
		t.Errorf("ParseYear() = %d, want 2013", year)
	}
}

// TestHarmoniaMundiParser_ParseCatalogNumber tests catalog extraction
func TestHarmoniaMundiParser_ParseCatalogNumber(t *testing.T) {
	html := `<div class="feature ref">HMC902170</div>`

	parser := NewHarmoniaMundiParser()
	catalog, err := parser.ParseCatalogNumber(html)

	if err != nil {
		t.Fatalf("ParseCatalogNumber() error = %v", err)
	}

	if catalog != "HMC902170" {
		t.Errorf("ParseCatalogNumber() = %q, want %q", catalog, "HMC902170")
	}
}

// TestHarmoniaMundiParser_ParseTracks tests track listing extraction
func TestHarmoniaMundiParser_ParseTracks(t *testing.T) {
	// Simplified track listing HTML
	html := `
		FELIX MENDELSSOHN BARTHOLDY [1809-1847]<br>
		· <b>Frohlocket, ihr Völker auf Erden, op.79/1</b> (1'38)<br>
		UWE GRONOSTAY<br>
		· <b>Die Nacht ist vorgedrungen</b> (2'26)<br>
		JOHANNES ECCARD<br>
		· <b>Ich lag in tiefster Todesnacht</b> (2'46)<br>
	`

	parser := NewHarmoniaMundiParser()
	tracks, err := parser.ParseTracks(html)

	if err != nil {
		t.Fatalf("ParseTracks() error = %v", err)
	}

	if len(tracks) != 3 {
		t.Fatalf("ParseTracks() returned %d tracks, want 3", len(tracks))
	}

	// Check first track
	track1 := tracks[0]
	if track1.Composer != "Felix Mendelssohn Bartholdy" {
		t.Errorf("Track 1 composer = %q, want %q", track1.Composer, "Felix Mendelssohn Bartholdy")
	}

	if track1.Title != "Frohlocket, ihr Völker auf Erden, op.79/1" {
		t.Errorf("Track 1 title = %q, want %q", track1.Title, "Frohlocket, ihr Völker auf Erden, op.79/1")
	}

	if track1.Track != 1 {
		t.Errorf("Track 1 number = %d, want 1", track1.Track)
	}

	if track1.Disc != 1 {
		t.Errorf("Track 1 disc = %d, want 1", track1.Disc)
	}

	// Check second track
	track2 := tracks[1]
	if track2.Composer != "Uwe Gronostay" {
		t.Errorf("Track 2 composer = %q, want %q", track2.Composer, "Uwe Gronostay")
	}

	// Check third track
	track3 := tracks[2]
	if track3.Composer != "Johannes Eccard" {
		t.Errorf("Track 3 composer = %q, want %q", track3.Composer, "Johannes Eccard")
	}
}

// TestHarmoniaMundiParser_ParseArtists tests artist extraction from byArtist field
func TestHarmoniaMundiParser_ParseArtists(t *testing.T) {
	html := `"byArtist": {"@type": "MusicGroup", "name": "RIAS Kammerchor, Hans-Christoph Rademann"}`

	parser := NewHarmoniaMundiParser()
	artistText, err := parser.ParseArtists(html)

	if err != nil {
		t.Fatalf("ParseArtists() error = %v", err)
	}

	expected := "RIAS Kammerchor, Hans-Christoph Rademann"
	if artistText != expected {
		t.Errorf("ParseArtists() = %q, want %q", artistText, expected)
	}
}

// TestHarmoniaMundiParser_FullParse tests complete album parsing
func TestHarmoniaMundiParser_FullParse(t *testing.T) {
	// Use the actual HTML file if available
	htmlFile := "/mnt/project/Noe_l___Weihnachten___Christmas____harmonia_mundi.html"

	htmlBytes, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Skip("HTML test file not available")
	}

	html := string(htmlBytes)
	parser := NewHarmoniaMundiParser()

	result, err := parser.Parse(html)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check album data
	data := result.Data()

	if data.Title == "" || data.Title == MissingTitle {
		t.Error("Title was not extracted")
	}

	if data.OriginalYear == 0 {
		t.Error("Year was not extracted")
	}

	if data.Edition == nil {
		t.Error("Edition was not extracted")
	} else {
		if data.Edition.Label != "harmonia mundi" {
			t.Errorf("Label = %q, want %q", data.Edition.Label, "harmonia mundi")
		}

		if data.Edition.CatalogNumber != "HMC902170" {
			t.Errorf("Catalog = %q, want %q", data.Edition.CatalogNumber, "HMC902170")
		}
	}

	if len(data.Tracks) == 0 {
		t.Error("No tracks were extracted")
	}

	// Check for parsing notes
	if result.ParsingNotes() == nil {
		t.Error("Expected parsing notes")
	}

	// Should have artist inferences
	if notes := result.ParsingNotes(); notes != nil {
		if _, ok := notes["artists"]; !ok {
			t.Error("Expected artist inference notes")
		}
	}
}

// TestHarmoniaMundiParser_ErrorHandling tests error collection
func TestHarmoniaMundiParser_ErrorHandling(t *testing.T) {
	// HTML with missing required fields
	html := `<html><body><div>No metadata here</div></body></html>`

	parser := NewHarmoniaMundiParser()
	result, err := parser.Parse(html)

	if err != nil {
		t.Fatalf("Parse() should not return error, got %v", err)
	}

	// Should have collected errors for missing fields
	if !result.HasErrors() {
		t.Error("Expected errors for missing required fields")
	}

	if !result.HasRequiredErrors() {
		t.Error("Expected required field errors")
	}

	// Check that sentinel values were used
	data := result.Data()
	if data.Title != MissingTitle {
		t.Errorf("Expected MissingTitle sentinel, got %q", data.Title)
	}

	if data.OriginalYear != MissingYear {
		t.Errorf("Expected MissingYear sentinel, got %d", data.OriginalYear)
	}
}

// TestHarmoniaMundiParser_MultiDisc tests multi-disc detection
func TestHarmoniaMundiParser_MultiDisc(t *testing.T) {
	t.Skip("Multi-disc test - need multi-disc example HTML")

	// TODO: Add test when we have a multi-disc HTML example
}
