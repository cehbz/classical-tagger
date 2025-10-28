package scraping

import (
	"os"
	"strings"
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestDiscogsParser_Parse(t *testing.T) {
	// Read test HTML file
	html, err := os.ReadFile("testdata/discogs_christmas.html")
	if err != nil {
		t.Skipf("Test HTML file not available: %v", err)
	}

	parser := NewDiscogsParser()
	result, err := parser.Parse(string(html))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil result")
	}

	data := result.Album

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

	// Test label extraction
	if data.Edition == nil || data.Edition.Label == "" {
		t.Error("Label not extracted")
	}

	// Test tracks extraction
	if len(data.Tracks) == 0 {
		t.Error("No tracks extracted")
	}

	// NEW: Verify album-level performers are present on ALL tracks
	for i, track := range data.Tracks {
		if len(track.Composers()) == 0 {
			t.Errorf("Track %d has no composer", i+1)
		}
		
		// Should have ensemble (RIAS Kammerchor)
		hasEnsemble := false
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleEnsemble {
				hasEnsemble = true
				if !strings.Contains(artist.Name, "RIAS") && !strings.Contains(artist.Name, "Kammerchor") {
					t.Errorf("Track %d: Expected RIAS Kammerchor in ensemble, got %q", i+1, artist.Name)
				}
				break
			}
		}
		if !hasEnsemble {
			t.Errorf("Track %d: Missing ensemble (RIAS Kammerchor)", i+1)
		}
		
		// Should have conductor (Hans-Christoph Rademann)
		hasConductor := false
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleConductor {
				hasConductor = true
				if !strings.Contains(artist.Name, "Rademann") {
					t.Errorf("Track %d: Expected Hans-Christoph Rademann as conductor, got %q", i+1, artist.Name)
				}
				break
			}
		}
		if !hasConductor {
			t.Errorf("Track %d: Missing conductor (Hans-Christoph Rademann)", i+1)
		}
		
		// Verify track number
		if track.Track != i+1 {
			t.Errorf("Track %d has wrong track number: got %d", i+1, track.Track)
		}
	}
}

func TestDiscogsParser_ParsePerformers(t *testing.T) {
	// Test with real Christmas album HTML
	html, err := os.ReadFile("testdata/discogs_christmas.html")
	if err != nil {
		t.Skipf("Test HTML file not available: %v", err)
	}

	parser := NewDiscogsParser()
	performers, err := parser.ParsePerformers(string(html))

	if err != nil {
		t.Fatalf("ParsePerformers() error = %v", err)
	}

	if len(performers) == 0 {
		t.Fatal("ParsePerformers() returned no performers")
	}

	// Should find RIAS Kammerchor and Hans-Christoph Rademann
	foundEnsemble := false
	foundConductor := false

	for _, performer := range performers {
		t.Logf("Performer: %s (role: %s)", performer.Name, performer.Role)
		
		if strings.Contains(performer.Name, "RIAS") || strings.Contains(performer.Name, "Kammerchor") {
			foundEnsemble = true
			if performer.Role != domain.RoleEnsemble {
				t.Errorf("RIAS Kammerchor has wrong role: got %s, want %s", performer.Role, domain.RoleEnsemble)
			}
		}
		
		if strings.Contains(performer.Name, "Rademann") {
			foundConductor = true
			// Should be Conductor (from "Chorus Master" role in releaseCredits)
			if performer.Role != domain.RoleConductor {
				t.Errorf("Rademann has wrong role: got %s, want %s", performer.Role, domain.RoleConductor)
			}
		}
	}

	if !foundEnsemble {
		t.Error("ParsePerformers() did not find ensemble (RIAS Kammerchor)")
	}
	if !foundConductor {
		t.Error("ParsePerformers() did not find conductor (Hans-Christoph Rademann)")
	}
}

func TestDiscogsParser_ParsePerformers_Simple(t *testing.T) {
	html := `
	<script type="application/ld+json" id="release_schema">
	{
		"@context":"http://schema.org",
		"@type":"MusicRelease",
		"name":"Test Album",
		"datePublished":2013,
		"releaseOf":{
			"@type":"MusicAlbum",
			"byArtist":[
				{"@type":"MusicGroup","name":"Berlin Philharmonic Orchestra"},
				{"@type":"Person","name":"Herbert von Karajan"}
			]
		}
	}
	</script>
	`

	parser := NewDiscogsParser()
	performers, err := parser.ParsePerformers(html)

	if err != nil {
		t.Fatalf("ParsePerformers() error = %v", err)
	}

	if len(performers) != 2 {
		t.Fatalf("ParsePerformers() returned %d performers, want 2", len(performers))
	}

	// Check ensemble
	found := false
	for _, p := range performers {
		if strings.Contains(p.Name, "Philharmonic") {
			found = true
			if p.Role != domain.RoleEnsemble {
				t.Errorf("Orchestra role = %s, want %s", p.Role, domain.RoleEnsemble)
			}
		}
	}
	if !found {
		t.Error("Did not find Berlin Philharmonic Orchestra")
	}

	// Check conductor (by name inference since no role in JSON-LD byArtist)
	found = false
	for _, p := range performers {
		if strings.Contains(p.Name, "Karajan") {
			found = true
			// Role will be inferred as Soloist (no "conductor" in name)
			if p.Role != domain.RoleSoloist {
				t.Errorf("Karajan role = %s, want %s (inferred from name)", p.Role, domain.RoleSoloist)
			}
		}
	}
	if !found {
		t.Error("Did not find Herbert von Karajan")
	}
}

func TestMapDiscogsRoleToDomainRole(t *testing.T) {
	tests := []struct {
		name         string
		discogsRole  string
		wantRole     domain.Role
		wantMappable bool
	}{
		// Ensemble roles
		{"choir", "Choir", domain.RoleEnsemble, true},
		{"chorus", "Chorus", domain.RoleEnsemble, true},
		{"orchestra", "Orchestra", domain.RoleEnsemble, true},
		{"ensemble", "Ensemble", domain.RoleEnsemble, true},
		{"kammerchor", "Kammerchor", domain.RoleEnsemble, true},
		{"vocal ensemble", "Vocal Ensemble", domain.RoleEnsemble, true},

		// Conductor roles
		{"conductor", "Conductor", domain.RoleConductor, true},
		{"chorus master", "Chorus Master", domain.RoleConductor, true},
		{"chorusmaster", "ChorusMaster", domain.RoleConductor, true},
		{"director", "Director", domain.RoleConductor, true},

		// Soloist roles
		{"soloist", "Soloist", domain.RoleSoloist, true},
		{"vocalist", "Vocalist", domain.RoleSoloist, true},
		{"singer", "Singer", domain.RoleSoloist, true},

		// Unmappable
		{"unknown", "Percussion", domain.RoleUnknown, false},
		{"empty", "", domain.RoleUnknown, false},
		{"random", "FooBar", domain.RoleUnknown, false},
		
		// Case insensitive
		{"uppercase", "CONDUCTOR", domain.RoleConductor, true},
		{"mixed case", "ChOiR", domain.RoleEnsemble, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRole, gotMappable := mapDiscogsRoleToDomainRole(tt.discogsRole)
			
			if gotMappable != tt.wantMappable {
				t.Errorf("mapDiscogsRoleToDomainRole(%q) mappable = %v, want %v",
					tt.discogsRole, gotMappable, tt.wantMappable)
			}
			
			if gotMappable && gotRole != tt.wantRole {
				t.Errorf("mapDiscogsRoleToDomainRole(%q) role = %v, want %v",
					tt.discogsRole, gotRole, tt.wantRole)
			}
		})
	}
}

func TestInferRoleFromName(t *testing.T) {
	tests := []struct {
		name     string
		artist   string
		wantRole domain.Role
	}{
		// Ensemble indicators
		{"philharmonic", "Berlin Philharmonic Orchestra", domain.RoleEnsemble},
		{"symphony", "London Symphony Orchestra", domain.RoleEnsemble},
		{"choir", "Westminster Choir", domain.RoleEnsemble},
		{"chorus", "Russian State Chorus", domain.RoleEnsemble},
		{"kammerchor", "RIAS Kammerchor", domain.RoleEnsemble},
		{"quartet", "Emerson String Quartet", domain.RoleEnsemble},
		{"ensemble", "Academy of Ancient Music Ensemble", domain.RoleEnsemble},
		{"chamber", "Chamber Orchestra of Europe", domain.RoleEnsemble},

		// Conductor indicators
		{"conductor explicit", "John Smith, conductor", domain.RoleConductor},
		{"director", "Music Director John Smith", domain.RoleConductor},

		// Soloist (default for individuals)
		{"individual name", "Martha Argerich", domain.RoleSoloist},
		{"pianist", "Glenn Gould", domain.RoleSoloist},
		{"violinist", "Anne-Sophie Mutter", domain.RoleSoloist},
		{"full name", "Hans-Christoph Rademann", domain.RoleSoloist},
		
		// Edge cases
		{"empty", "", domain.RoleSoloist},
		{"single word", "Madonna", domain.RoleSoloist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRole := inferRoleFromName(tt.artist)
			if gotRole != tt.wantRole {
				t.Errorf("inferRoleFromName(%q) = %v, want %v",
					tt.artist, gotRole, tt.wantRole)
			}
		})
	}
}

func TestDiscogsParser_ParseFromJSONLD(t *testing.T) {
	html := `
	<script type="application/ld+json" id="release_schema">
	{
		"@context":"http://schema.org",
		"@type":"MusicRelease",
		"name":"Noël! Christmas! Weihnachten!",
		"datePublished":2013,
		"catalogNumber":"HMC 902170",
		"recordLabel":[{
			"@type":"Organization",
			"name":"Test Label"
		}]
	}
	</script>
	`

	parser := NewDiscogsParser()

	// Test title
	title, err := parser.ParseTitle(html)
	if err != nil {
		t.Errorf("ParseTitle() error = %v", err)
	}
	if title != "Noël! Christmas! Weihnachten!" {
		t.Errorf("ParseTitle() = %q, want %q", title, "Noël! Christmas! Weihnachten!")
	}

	// Test year
	year, err := parser.ParseYear(html)
	if err != nil {
		t.Errorf("ParseYear() error = %v", err)
	}
	if year != 2013 {
		t.Errorf("ParseYear() = %d, want 2013", year)
	}

	// Test catalog
	catalog, err := parser.ParseCatalogNumber(html)
	if err != nil {
		t.Errorf("ParseCatalogNumber() error = %v", err)
	}
	if catalog != "HMC 902170" {
		t.Errorf("ParseCatalogNumber() = %q, want %q", catalog, "HMC 902170")
	}

	// Test label
	label, err := parser.ParseLabel(html)
	if err != nil {
		t.Errorf("ParseLabel() error = %v", err)
	}
	if label != "Test Label" {
		t.Errorf("ParseLabel() = %q, want %q", label, "Test Label")
	}
}

func TestDiscogsParser_ParseTracks(t *testing.T) {
	html := `
	<table class="tracklist_ZdQ0I">
		<tbody>
			<tr data-track-position="1">
				<td class="trackPos_n8vad">1</td>
				<td class="trackTitle_loyWF">
					<span>Frohlocket, Ihr Völker Auf Erden (op.79/1)</span>
					<div class="credits_vzBtg">
						<span>Composed By</span> – 
						<a href="/artist/623293-Felix-Mendelssohn-Bartholdy">Felix Mendelssohn Bartholdy</a>
					</div>
				</td>
				<td class="duration_GhhxK">1:38</td>
			</tr>
			<tr data-track-position="2">
				<td class="trackPos_n8vad">2</td>
				<td class="trackTitle_loyWF">
					<span>Die Nacht Ist Vorgedrungen</span>
					<div class="credits_vzBtg">
						<span>Composed By</span> – 
						<a href="/artist/837343-Uwe-Gronostay">Uwe Gronostay</a>
					</div>
				</td>
				<td class="duration_GhhxK">2:26</td>
			</tr>
		</tbody>
	</table>
	`

	parser := NewDiscogsParser()
	tracks, err := parser.ParseTracks(html)

	if err != nil {
		t.Fatalf("ParseTracks() error = %v", err)
	}

	if len(tracks) != 2 {
		t.Fatalf("ParseTracks() got %d tracks, want 2", len(tracks))
	}

	// Check first track
	if tracks[0].Title != "Frohlocket, Ihr Völker Auf Erden (op.79/1)" {
		t.Errorf("Track 1 title = %q", tracks[0].Title)
	}
	if composers := tracks[0].Composers(); len(composers) == 0 || composers[0].Name != "Felix Mendelssohn Bartholdy" {
		composerName := "<none>"
		if len(composers) > 0 {
			composerName = composers[0].Name
		}
		t.Errorf("Track 1 composer = %q", composerName)
	}

	// Check second track
	if tracks[1].Title != "Die Nacht Ist Vorgedrungen" {
		t.Errorf("Track 2 title = %q", tracks[1].Title)
	}
	if composers := tracks[1].Composers(); len(composers) == 0 || composers[0].Name != "Uwe Gronostay" {
		composerName := "<none>"
		if len(composers) > 0 {
			composerName = composers[0].Name
		}
		t.Errorf("Track 2 composer = %q", composerName)
	}
}

// TestDiscogsParser_ParseTracks_NoDuplicateComposers tests the bug where composer names
// are duplicated in the output.
func TestDiscogsParser_ParseTracks_NoDuplicateComposers(t *testing.T) {
	html := `
	<html>
	<head>
		<script type="application/ld+json" id="release_schema">
		{
			"@context":"http://schema.org",
			"@type":"MusicRelease",
			"name":"Test Album",
			"datePublished":2013
		}
		</script>
	</head>
	<body>
		<table class="tracklist_ZdQ0I">
			<tbody>
				<tr data-track-position="1">
					<td class="trackPos_n8vad">1</td>
					<td class="trackTitle_loyWF">
						<span>Frohlocket, Ihr Völker Auf Erden (op.79/1)</span>
						<div class="credits_vzBtg">
							<span>Composed By</span> – 
							<a href="/artist/623293-Felix-Mendelssohn-Bartholdy">Felix Mendelssohn Bartholdy</a>
						</div>
					</td>
					<td class="duration_GhhxK">1:38</td>
				</tr>
				<tr data-track-position="2">
					<td class="trackPos_n8vad">2</td>
					<td class="trackTitle_loyWF">
						<span>Die Nacht Ist Vorgedrungen</span>
						<div class="credits_vzBtg">
							<span>Composed By</span> – 
							<a href="/artist/837343-Uwe-Gronostay">Uwe Gronostay</a>
						</div>
					</td>
					<td class="duration_GhhxK">2:26</td>
				</tr>
				<tr data-track-position="3">
					<td class="trackPos_n8vad">3</td>
					<td class="trackTitle_loyWF">
						<span>Ave Maria</span>
						<div class="credits_vzBtg">
							<span>Composed By</span> – 
							<a href="/artist/25228-Anton-Bruckner">Anton Bruckner</a>
						</div>
					</td>
					<td class="duration_GhhxK">4:12</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	parser := NewDiscogsParser()
	result, err := parser.Parse(html)

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tracks := result.Album.Tracks
	if len(tracks) != 3 {
		t.Fatalf("Expected 3 tracks, got %d", len(tracks))
	}

	// Test each track's composer for duplication
	expectedComposers := []string{
		"Felix Mendelssohn Bartholdy",
		"Uwe Gronostay",
		"Anton Bruckner",
	}

	for i, track := range tracks {
		var composer string
		composers := track.Composers()
		switch len(composers) {
		case 0:
			t.Errorf("Track %d has no composer", i+1)
		case 1:
			composer = composers[0].Name
		default:
			t.Errorf("Track %d has multiple composers: %d", i+1, len(composers))
		}
		expected := expectedComposers[i]

		// Check if composer name is duplicated (exact concatenation)
		if len(composer) >= len(expected)*2 && composer == expected+expected {
			t.Errorf("Track %d: composer is exactly duplicated: %q", i+1, composer)
			t.Logf("BUG: Composer name appears twice concatenated")
			t.Logf("Expected: %q", expected)
		}

		// Check for partial duplication patterns
		if composer != expected && strings.Contains(composer, expected) {
			// Count occurrences
			count := strings.Count(composer, expected)
			if count > 1 {
				t.Errorf("Track %d: composer name appears %d times in %q", i+1, count, composer)
			}
		}

		// Check individual words for duplication
		words := strings.Fields(composer)
		wordCount := make(map[string]int)
		for _, word := range words {
			wordCount[word]++
			if wordCount[word] > 1 && len(word) > 3 { // Ignore short words
				t.Errorf("Track %d: word %q appears %d times in composer %q",
					i+1, word, wordCount[word], composer)
			}
		}

		// Final check: composer should match expected exactly
		if composer != expected {
			t.Errorf("Track %d: composer = %q, want %q", i+1, composer, expected)
		}
	}
}

// TestDiscogsParser_ParseTracks_MultiMovementWork tests that movement tracks
// include the parent work name in their title, matching the behavior of PrestoParser.
// Rule reference: Movement tracks of multi-movement works should include the work name.
func TestDiscogsParser_ParseTracks_MultiMovementWork(t *testing.T) {
	html := `
	<table class="tracklist_ZdQ0I">
		<tbody>
			<!-- Regular standalone track -->
			<tr data-track-position="15">
				<td class="trackPos_n8vad">15</td>
				<td class="trackTitle_loyWF">
					<span>In Dulci Jubilo</span>
					<div class="credits_vzBtg">
						<span>Composed By</span> – 
						<a href="/artist/856233-Michael-Praetorius">Michael Praetorius</a>
					</div>
				</td>
			</tr>
			
			<!-- Multi-movement work heading -->
			<tr class="heading_mkZNt">
				<td></td>
				<td class="trackTitle_loyWF">Quatre Motets Pour Le Temps de Noël
					<div class="credits_vzBtg">
						<span>Composed By</span> – 
						<a href="/artist/361814-Francis-Poulenc">Francis Poulenc</a>
					</div>
				</td>
			</tr>
			
			<!-- Movement 1 -->
			<tr class="subtrack_o3GgI">
				<td class="subtrackPos_HC1me">16</td>
				<td class="trackTitle_loyWF">
					<span>O Magnum Mysterium</span>
				</td>
			</tr>
			
			<!-- Movement 2 -->
			<tr class="subtrack_o3GgI">
				<td class="subtrackPos_HC1me">17</td>
				<td class="trackTitle_loyWF">
					<span>Quem Vidistis Pastores Dicite</span>
				</td>
			</tr>
			
			<!-- Movement 3 -->
			<tr class="subtrack_o3GgI">
				<td class="subtrackPos_HC1me">18</td>
				<td class="trackTitle_loyWF">
					<span>Videntes Stellam</span>
				</td>
			</tr>
			
			<!-- Movement 4 -->
			<tr class="subtrack_o3GgI">
				<td class="subtrackPos_HC1me">19</td>
				<td class="trackTitle_loyWF">
					<span>Hodie Christus Natus Est</span>
				</td>
			</tr>
			
			<!-- Next standalone track -->
			<tr data-track-position="20">
				<td class="trackPos_n8vad">20</td>
				<td class="trackTitle_loyWF">
					<span>Stille Nacht</span>
					<div class="credits_vzBtg">
						<span>Composed By</span> – 
						<a href="/artist/1316922-Franz-Xaver-Gruber">Franz Xaver Gruber</a>
					</div>
				</td>
			</tr>
		</tbody>
	</table>
	`

	parser := NewDiscogsParser()
	tracks, err := parser.ParseTracks(html)

	if err != nil {
		t.Fatalf("ParseTracks() error = %v", err)
	}

	// Should have 6 tracks total (1 regular + 4 Poulenc movements + 1 final)
	if len(tracks) != 6 {
		t.Errorf("Got %d tracks, want 6", len(tracks))
		t.Logf("Tracks extracted:")
		for _, track := range tracks {
			composers := track.Composers()
			composerName := "<none>"
			if len(composers) > 0 {
				composerName = composers[0].Name
			}
			t.Logf("  %d. %s (composer: %s)", track.Track, track.Title, composerName)
		}
	}

	// Verify track 1 is the regular track before the multi-movement work
	if len(tracks) >= 1 {
		track := tracks[0]
		if !strings.Contains(track.Title, "In Dulci Jubilo") {
			t.Errorf("Track 1 title = %q, want to contain 'In Dulci Jubilo'", track.Title)
		}
		if composers := track.Composers(); len(composers) == 0 || composers[0].Name != "Michael Praetorius" {
			composerName := "<none>"
			if len(composers) > 0 {
				composerName = composers[0].Name
			}
			t.Errorf("Track 1 composer = %q, want 'Michael Praetorius'", composerName)
		}
		if track.Track != 15 {
			t.Errorf("Track 1 track number = %d, want 15", track.Track)
		}
	}

	// Check that Poulenc movements have cycle name prepended
	expectedPoulencTitles := []string{
		"Quatre Motets Pour Le Temps de Noël: O Magnum Mysterium",
		"Quatre Motets Pour Le Temps de Noël: Quem Vidistis Pastores Dicite",
		"Quatre Motets Pour Le Temps de Noël: Videntes Stellam",
		"Quatre Motets Pour Le Temps de Noël: Hodie Christus Natus Est",
	}

	for i := 0; i < 4 && i+1 < len(tracks); i++ {
		track := tracks[i+1] // Skip first regular track
		expected := expectedPoulencTitles[i]

		if track.Title != expected {
			t.Errorf("Track %d title = %q, want %q", track.Track, track.Title, expected)
		}

		// Verify composer is preserved from parent heading
		if composers := track.Composers(); len(composers) == 0 || composers[0].Name != "Francis Poulenc" {
			composerName := "<none>"
			if len(composers) > 0 {
				composerName = composers[0].Name
			}
			t.Errorf("Track %d composer = %q, want 'Francis Poulenc'", track.Track, composerName)
		}
	}

	// Verify last track is standalone work (not part of Poulenc cycle)
	if len(tracks) >= 6 {
		lastTrack := tracks[5]
		if !strings.Contains(lastTrack.Title, "Stille Nacht") {
			t.Errorf("Track 6 should contain 'Stille Nacht', got %q", lastTrack.Title)
		}
		if composers := lastTrack.Composers(); len(composers) == 0 || composers[0].Name != "Franz Xaver Gruber" {
			composerName := "<none>"
			if len(composers) > 0 {
				composerName = composers[0].Name
			}
			t.Errorf("Track 6 composer = %q, want 'Franz Xaver Gruber'", composerName)
		}
	}
}