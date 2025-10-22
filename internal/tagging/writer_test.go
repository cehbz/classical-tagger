package tagging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ============================================================================
// Phase 1: Conversion Logic Tests (No FLAC files needed)
// ============================================================================

func TestTrackToTagData_Basic(t *testing.T) {
	// Create test album
	album, _ := domain.NewAlbum("Goldberg Variations", 1981)
	edition, _ := domain.NewEdition("Sony Classical", 1981)
	edition = edition.WithCatalogNumber("SMK89245")
	album = album.WithEdition(edition)

	// Create test track
	composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	performer, _ := domain.NewArtist("Glenn Gould", domain.RoleSoloist)
	track, _ := domain.NewTrack(1, 1, "Aria", []domain.Artist{composer, performer})

	// Convert to tagData
	tags := trackToTagData(track, album)

	// Verify track-level fields
	if tags.title != "Aria" {
		t.Errorf("title = %q, want %q", tags.title, "Aria")
	}
	if tags.composer != "Johann Sebastian Bach" {
		t.Errorf("composer = %q, want %q", tags.composer, "Johann Sebastian Bach")
	}
	if tags.artist != "Glenn Gould" {
		t.Errorf("artist = %q, want %q (composer should be excluded)", tags.artist, "Glenn Gould")
	}
	if tags.trackNumber != "1" {
		t.Errorf("trackNumber = %q, want %q", tags.trackNumber, "1")
	}
	if tags.discNumber != "1" {
		t.Errorf("discNumber = %q, want %q", tags.discNumber, "1")
	}

	// Verify album-level fields
	if tags.album != "Goldberg Variations" {
		t.Errorf("album = %q, want %q", tags.album, "Goldberg Variations")
	}
	if tags.date != "1981" {
		t.Errorf("date = %q, want %q", tags.date, "1981")
	}
	if tags.originalDate != "1981" {
		t.Errorf("originalDate = %q, want %q", tags.originalDate, "1981")
	}
	if tags.label != "Sony Classical" {
		t.Errorf("label = %q, want %q", tags.label, "Sony Classical")
	}
	if tags.catalogNumber != "SMK89245" {
		t.Errorf("catalogNumber = %q, want %q", tags.catalogNumber, "SMK89245")
	}
}

func TestTrackToTagData_MultiplePerformers(t *testing.T) {
	album, _ := domain.NewAlbum("Piano Concerto No. 1", 1976)

	composer, _ := domain.NewArtist("Ludwig van Beethoven", domain.RoleComposer)
	soloist, _ := domain.NewArtist("Martha Argerich", domain.RoleSoloist)
	ensemble, _ := domain.NewArtist("London Symphony Orchestra", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Claudio Abbado", domain.RoleConductor)

	track, _ := domain.NewTrack(1, 1, "Piano Concerto No. 1: I. Allegro",
		[]domain.Artist{composer, soloist, ensemble, conductor})

	tags := trackToTagData(track, album)

	// Artist field should be "Soloist, Ensemble, Conductor" (no composer)
	expected := "Martha Argerich, London Symphony Orchestra, Claudio Abbado"
	if tags.artist != expected {
		t.Errorf("artist = %q, want %q", tags.artist, expected)
	}

	// Composer should be separate
	if tags.composer != "Ludwig van Beethoven" {
		t.Errorf("composer = %q, want %q", tags.composer, "Ludwig van Beethoven")
	}
}

func TestTrackToTagData_NoEdition(t *testing.T) {
	album, _ := domain.NewAlbum("Test Album", 2020)
	composer, _ := domain.NewArtist("Test Composer", domain.RoleComposer)
	track, _ := domain.NewTrack(1, 1, "Test Track", []domain.Artist{composer})

	tags := trackToTagData(track, album)

	// Edition fields should be empty
	if tags.label != "" {
		t.Errorf("label = %q, want empty", tags.label)
	}
	if tags.catalogNumber != "" {
		t.Errorf("catalogNumber = %q, want empty", tags.catalogNumber)
	}
	// Date should still be set to original year
	if tags.date != "2020" {
		t.Errorf("date = %q, want %q", tags.date, "2020")
	}
	if tags.originalDate != "2020" {
		t.Errorf("originalDate = %q, want %q", tags.originalDate, "2020")
	}
}

func TestTrackToTagData_DifferentEditionYear(t *testing.T) {
	album, _ := domain.NewAlbum("Test Album", 1960)
	edition, _ := domain.NewEdition("Reissue Label", 2020)
	album = album.WithEdition(edition)

	composer, _ := domain.NewArtist("Test Composer", domain.RoleComposer)
	track, _ := domain.NewTrack(1, 1, "Test Track", []domain.Artist{composer})

	tags := trackToTagData(track, album)

	// DATE should be edition year, ORIGINALDATE should be original
	if tags.date != "2020" {
		t.Errorf("date = %q, want %q (edition year)", tags.date, "2020")
	}
	if tags.originalDate != "1960" {
		t.Errorf("originalDate = %q, want %q (original year)", tags.originalDate, "1960")
	}
}

func TestFormatArtists_RoleOrdering(t *testing.T) {
	tests := []struct {
		name     string
		artists  []domain.Artist
		expected string
	}{
		{
			name: "soloist only",
			artists: []domain.Artist{
				mustArtist("Glenn Gould", domain.RoleSoloist),
			},
			expected: "Glenn Gould",
		},
		{
			name: "ensemble only",
			artists: []domain.Artist{
				mustArtist("Berlin Philharmonic", domain.RoleEnsemble),
			},
			expected: "Berlin Philharmonic",
		},
		{
			name: "conductor only",
			artists: []domain.Artist{
				mustArtist("Herbert von Karajan", domain.RoleConductor),
			},
			expected: "Herbert von Karajan",
		},
		{
			name: "soloist and ensemble",
			artists: []domain.Artist{
				mustArtist("Martha Argerich", domain.RoleSoloist),
				mustArtist("LSO", domain.RoleEnsemble),
			},
			expected: "Martha Argerich, LSO",
		},
		{
			name: "all three roles",
			artists: []domain.Artist{
				mustArtist("Kyung-Wha Chung", domain.RoleSoloist),
				mustArtist("Vienna Philharmonic", domain.RoleEnsemble),
				mustArtist("Simon Rattle", domain.RoleConductor),
			},
			expected: "Kyung-Wha Chung, Vienna Philharmonic, Simon Rattle",
		},
		{
			name: "multiple soloists",
			artists: []domain.Artist{
				mustArtist("Peter Seiffert", domain.RoleSoloist),
				mustArtist("Thomas Hampson", domain.RoleSoloist),
				mustArtist("City of Birmingham Symphony Orchestra", domain.RoleEnsemble),
				mustArtist("Simon Rattle", domain.RoleConductor),
			},
			expected: "Peter Seiffert, Thomas Hampson, City of Birmingham Symphony Orchestra, Simon Rattle",
		},
		{
			name: "with composer (should be excluded)",
			artists: []domain.Artist{
				mustArtist("Johann Sebastian Bach", domain.RoleComposer),
				mustArtist("Glenn Gould", domain.RoleSoloist),
			},
			expected: "Glenn Gould",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatArtists(tt.artists)
			if result != tt.expected {
				t.Errorf("formatArtists() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDetermineAlbumArtist_SingleUniversal(t *testing.T) {
	album, _ := domain.NewAlbum("Test Album", 2020)

	composer1, _ := domain.NewArtist("Composer 1", domain.RoleComposer)
	composer2, _ := domain.NewArtist("Composer 2", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Test Ensemble", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Test Conductor", domain.RoleConductor)

	// Track 1
	track1, _ := domain.NewTrack(1, 1, "Track 1", []domain.Artist{composer1, ensemble, conductor})
	album.AddTrack(track1)

	// Track 2 - same performers, different composer
	track2, _ := domain.NewTrack(1, 2, "Track 2", []domain.Artist{composer2, ensemble, conductor})
	album.AddTrack(track2)

	albumArtist, universal := determineAlbumArtist(album)

	expected := "Test Ensemble, Test Conductor"
	if albumArtist != expected {
		t.Errorf("albumArtist = %q, want %q", albumArtist, expected)
	}

	if len(universal) != 2 {
		t.Errorf("universal count = %d, want 2", len(universal))
	}
}

func TestDetermineAlbumArtist_NoUniversal(t *testing.T) {
	album, _ := domain.NewAlbum("Test Album", 2020)

	composer, _ := domain.NewArtist("Composer", domain.RoleComposer)
	soloist1, _ := domain.NewArtist("Soloist 1", domain.RoleSoloist)
	soloist2, _ := domain.NewArtist("Soloist 2", domain.RoleSoloist)

	// Track 1 - Soloist 1
	track1, _ := domain.NewTrack(1, 1, "Track 1", []domain.Artist{composer, soloist1})
	album.AddTrack(track1)

	// Track 2 - Soloist 2 (different)
	track2, _ := domain.NewTrack(1, 2, "Track 2", []domain.Artist{composer, soloist2})
	album.AddTrack(track2)

	albumArtist, universal := determineAlbumArtist(album)

	if albumArtist != "" {
		t.Errorf("albumArtist = %q, want empty (no universal performers)", albumArtist)
	}

	if len(universal) != 0 {
		t.Errorf("universal count = %d, want 0", len(universal))
	}
}

func TestDetermineAlbumArtist_HarmoniaMundial(t *testing.T) {
	// Real example from Harmonia Mundi album
	album, _ := domain.NewAlbum("Noël · Weihnachten · Christmas", 2013)

	mendelssohn, _ := domain.NewArtist("Felix Mendelssohn Bartholdy", domain.RoleComposer)
	brahms, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
	poulenc, _ := domain.NewArtist("Francis Poulenc", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("RIAS Kammerchor Berlin", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Hans-Christoph Rademann", domain.RoleConductor)

	// All tracks have same ensemble and conductor, different composers
	track1, _ := domain.NewTrack(1, 1, "Track 1", []domain.Artist{mendelssohn, ensemble, conductor})
	track2, _ := domain.NewTrack(1, 2, "Track 2", []domain.Artist{brahms, ensemble, conductor})
	track3, _ := domain.NewTrack(1, 3, "Track 3", []domain.Artist{poulenc, ensemble, conductor})

	album.AddTrack(track1)
	album.AddTrack(track2)
	album.AddTrack(track3)

	albumArtist, universal := determineAlbumArtist(album)

	// Should have both ensemble and conductor
	if !strings.Contains(albumArtist, "RIAS Kammerchor Berlin") {
		t.Errorf("albumArtist missing ensemble: %q", albumArtist)
	}
	if !strings.Contains(albumArtist, "Hans-Christoph Rademann") {
		t.Errorf("albumArtist missing conductor: %q", albumArtist)
	}

	if len(universal) != 2 {
		t.Errorf("universal count = %d, want 2 (ensemble + conductor)", len(universal))
	}
}

// ============================================================================
// Phase 2: Data Loss Detection Tests
// ============================================================================

func TestSplitArtists(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single artist",
			input:    "Glenn Gould",
			expected: []string{"Glenn Gould"},
		},
		{
			name:     "two artists comma",
			input:    "Glenn Gould, Herbert von Karajan",
			expected: []string{"Glenn Gould", "Herbert von Karajan"},
		},
		{
			name:     "three artists comma",
			input:    "Kyung-Wha Chung, Vienna Philharmonic, Simon Rattle",
			expected: []string{"Kyung-Wha Chung", "Vienna Philharmonic", "Simon Rattle"},
		},
		{
			name:     "semicolon separator",
			input:    "Artist 1; Artist 2; Artist 3",
			expected: []string{"Artist 1", "Artist 2", "Artist 3"},
		},
		{
			name:     "with extra spaces",
			input:    "Artist 1 ,  Artist 2  ,Artist 3",
			expected: []string{"Artist 1", "Artist 2", "Artist 3"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitArtists(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitArtists() count = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("splitArtists()[%d] = %q, want %q", i, result[i], expected)
				}
			}
		})
	}
}

func TestIsArtistSuperset_Safe(t *testing.T) {
	tests := []struct {
		name    string
		oldVal  string
		newVal  string
		wantOK  bool
		wantMsg string
	}{
		{
			name:   "identical",
			oldVal: "Glenn Gould",
			newVal: "Glenn Gould",
			wantOK: true,
		},
		{
			name:   "adding artist",
			oldVal: "Glenn Gould",
			newVal: "Glenn Gould, Vienna Philharmonic",
			wantOK: true,
		},
		{
			name:   "adding two artists",
			oldVal: "Glenn Gould",
			newVal: "Glenn Gould, Vienna Philharmonic, Herbert von Karajan",
			wantOK: true,
		},
		{
			name:   "reordering safe",
			oldVal: "Karajan, Berlin Philharmonic",
			newVal: "Berlin Philharmonic, Karajan",
			wantOK: true,
		},
		{
			name:    "removing artist",
			oldVal:  "Glenn Gould, Vienna Philharmonic, Herbert von Karajan",
			newVal:  "Glenn Gould",
			wantOK:  false,
			wantMsg: "Vienna Philharmonic",
		},
		{
			name:    "completely different",
			oldVal:  "Artist A",
			newVal:  "Artist B",
			wantOK:  false,
			wantMsg: "Artist A",
		},
		{
			name:   "case insensitive match",
			oldVal: "glenn gould",
			newVal: "Glenn Gould",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := isArtistSuperset(tt.newVal, tt.oldVal)
			if ok != tt.wantOK {
				t.Errorf("isArtistSuperset() ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK && !strings.Contains(msg, tt.wantMsg) {
				t.Errorf("isArtistSuperset() msg = %q, should contain %q", msg, tt.wantMsg)
			}
		})
	}
}

// ============================================================================
// Phase 3: FLAC Integration Tests (Skipped without fixture)
// ============================================================================

func TestFLACWriter_DryRun(t *testing.T) {
	writer := NewFLACWriter()
	writer.SetDryRun(true)

	album, _ := domain.NewAlbum("Test", 2020)
	composer, _ := domain.NewArtist("Test", domain.RoleComposer)
	track, _ := domain.NewTrack(1, 1, "Test", []domain.Artist{composer})

	// In dry-run, should not error even with non-existent file
	err := writer.WriteTrack("/nonexistent/path.flac", track, album)
	if err != nil {
		t.Errorf("DryRun should not error, got: %v", err)
	}
}

func TestFLACWriter_BackupRestore(t *testing.T) {
	writer := NewFLACWriter()

	// Create a test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.flac")
	originalContent := []byte("test flac data")
	err := os.WriteFile(testFile, originalContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Backup
	backupPath, err := writer.BackupFile(testFile)
	if err != nil {
		t.Fatalf("BackupFile() error = %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Modify original
	err = os.WriteFile(testFile, []byte("modified"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Restore
	err = writer.RestoreBackup(backupPath, testFile)
	if err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}

	// Verify restored content matches original
	restoredContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(restoredContent) != string(originalContent) {
		t.Error("Restored content does not match original")
	}
}

func TestFLACWriter_WriteTrack_Integration(t *testing.T) {
	t.Skip("Requires FLAC test fixture - implement after basic structure works")

	// This test will:
	// 1. Create or use a test FLAC file
	// 2. Write tags using WriteTrack
	// 3. Read back using FLACReader
	// 4. Verify all tags match
}

func TestFLACWriter_PreservesExistingTags(t *testing.T) {
	t.Skip("Requires FLAC test fixture")

	// This test will:
	// 1. Create FLAC with REPLAYGAIN_* tags
	// 2. Write music metadata
	// 3. Verify REPLAYGAIN_* tags still present
}

// ============================================================================
// Helper Functions
// ============================================================================

func mustArtist(name string, role domain.Role) domain.Artist {
	artist, err := domain.NewArtist(name, role)
	if err != nil {
		panic(err)
	}
	return artist
}
