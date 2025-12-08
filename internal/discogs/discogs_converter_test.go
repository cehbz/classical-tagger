package discogs

import (
	"strings"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestConvertDiscogsRelease_DeduplicateAlbumArtists(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Artists: []Artist{
			{Name: "RIAS-Kammerchor"},
			{Name: "Hans-Christoph Rademann"},
		},
		ExtraArtists: []Artist{
			{Name: "RIAS-Kammerchor", Role: "Choir"},
			{Name: "Hans-Christoph Rademann", Role: "Chorus Master"},
		},
		Tracklist: []Track{
			{Position: "1", Title: "Track 1"},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}
	if torrent == nil {
		t.Fatal("convertDiscogsRelease returned nil")
	}

	// Should have deduplicated artists - RIAS-Kammerchor should appear once with ensemble role
	// Hans-Christoph Rademann should appear once with conductor role (Chorus Master -> conductor)
	// Roles should come from ExtraArtists since main artists have no roles
	riasCount := 0
	hansCount := 0
	for _, artist := range torrent.AlbumArtist {
		if artist.Name == "RIAS-Kammerchor" {
			riasCount++
			if artist.Role != domain.RoleEnsemble {
				t.Errorf("RIAS-Kammerchor should have ensemble role (from extraartists), got %v", artist.Role)
			}
		}
		if artist.Name == "Hans-Christoph Rademann" {
			hansCount++
			if artist.Role != domain.RoleConductor {
				t.Errorf("Hans-Christoph Rademann should have conductor role (from extraartists), got %v", artist.Role)
			}
		}
	}

	if riasCount != 1 {
		t.Errorf("Expected RIAS-Kammerchor to appear once, got %d times", riasCount)
	}
	if hansCount != 1 {
		t.Errorf("Expected Hans-Christoph Rademann to appear once, got %d times", hansCount)
	}
}

func TestConvertDiscogsRelease_RoleFromExtraArtists(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Artists: []Artist{
			{Name: "RIAS-Kammerchor"}, // No role in main artists
		},
		ExtraArtists: []Artist{
			{Name: "RIAS-Kammerchor", Role: "Choir"}, // Role in extraartists
		},
		Tracklist: []Track{
			{Position: "1", Title: "Track 1"},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}

	// Should use role from extraartists
	found := false
	for _, artist := range torrent.AlbumArtist {
		if artist.Name == "RIAS-Kammerchor" {
			found = true
			if artist.Role != domain.RoleEnsemble {
				t.Errorf("RIAS-Kammerchor should have ensemble role from extraartists, got %v", artist.Role)
			}
		}
	}
	if !found {
		t.Error("RIAS-Kammerchor should be in album artists")
	}
}

func TestConvertDiscogsRelease_RoleFromLocalMetadata(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Artists: []Artist{
			{Name: "RIAS-Kammerchor"}, // No role
		},
		Tracklist: []Track{
			{Position: "1", Title: "Track 1"},
		},
	}

	// Local metadata has role information
	localTorrent := &domain.Torrent{
		AlbumArtist: []domain.Artist{
			{Name: "RIAS-Kammerchor", Role: domain.RoleEnsemble},
		},
	}

	torrent, err := release.DomainTorrent("test-path", localTorrent)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}

	// Should use role from local metadata
	found := false
	for _, artist := range torrent.AlbumArtist {
		if artist.Name == "RIAS-Kammerchor" {
			found = true
			if artist.Role != domain.RoleEnsemble {
				t.Errorf("RIAS-Kammerchor should have ensemble role from local metadata, got %v", artist.Role)
			}
		}
	}
	if !found {
		t.Error("RIAS-Kammerchor should be in album artists")
	}
}

func TestConvertDiscogsRelease_ErrorOnUnknownRole(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Artists: []Artist{
			{Name: "Unknown Artist"}, // No role, no extraartists match, no local metadata
		},
		Tracklist: []Track{
			{Position: "1", Title: "Track 1"},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err == nil {
		t.Error("Expected error for unknown role, got nil")
	}
	if torrent != nil {
		t.Error("Expected nil torrent when error occurs")
	}
	if err != nil && !strings.Contains(err.Error(), "cannot determine role") {
		t.Errorf("Expected error message about role determination, got: %v", err)
	}
}

func TestConvertDiscogsRelease_SkipParentWorkEntries(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Tracklist: []Track{
			{Position: "1", Title: "Track 1"},
			{
				Position: "",
				Title:    "Parent Work",
				Artists:  []Artist{{Name: "Composer", Role: "Composed By"}},
				SubTracks: []Track{
					{Position: "2", Title: "Movement 1"},
				},
			},
			{Position: "3", Title: "Track 3"},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}
	if torrent == nil {
		t.Fatal("convertDiscogsRelease returned nil")
	}

	// Should have 3 tracks: Track 1, Movement 1 (from subtrack), Track 3
	if len(torrent.Files) != 3 {
		t.Errorf("Expected 3 tracks, got %d", len(torrent.Files))
	}

	// Track 2 (from subtrack) should have parent composer and prepended title
	track2 := torrent.Tracks()[1]
	if len(track2.Artists) == 0 {
		t.Error("Track 2 should have composer from parent work")
	} else if track2.Artists[0].Name != "Composer" || track2.Artists[0].Role != domain.RoleComposer {
		t.Errorf("Track 2 should have composer 'Composer', got %v", track2.Artists[0])
	}
	if track2.Title != "Parent Work: Movement 1" {
		t.Errorf("Track 2 title should be 'Parent Work: Movement 1', got %q", track2.Title)
	}
}

func TestConvertDiscogsRelease_MultipleParentWorks(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Tracklist: []Track{
			{Position: "1", Title: "Track 1", Artists: []Artist{{Name: "Composer1", Role: "Composed By"}}},
			{
				Position: "",
				Title:    "Parent Work 1",
				Artists:  []Artist{{Name: "ParentComposer1", Role: "Composed By"}},
				SubTracks: []Track{
					{Position: "2", Title: "Movement 1"},
					{Position: "3", Title: "Movement 2"},
				},
			},
			{
				Position: "",
				Title:    "Parent Work 2",
				Artists:  []Artist{{Name: "ParentComposer2", Role: "Composed By"}},
				SubTracks: []Track{
					{Position: "4", Title: "Movement A"},
				},
			},
			{Position: "5", Title: "Track 5", Artists: []Artist{{Name: "Composer5", Role: "Composed By"}}},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}
	if torrent == nil {
		t.Fatal("convertDiscogsRelease returned nil")
	}

	tracks := torrent.Tracks()
	if len(tracks) != 5 {
		t.Fatalf("Expected 5 tracks, got %d", len(tracks))
	}

	// Track 1: Has its own composer
	if tracks[0].Artists[0].Name != "Composer1" {
		t.Errorf("Track 1 should have Composer1, got %s", tracks[0].Artists[0].Name)
	}

	// Track 2: From Parent Work 1
	if tracks[1].Artists[0].Name != "ParentComposer1" {
		t.Errorf("Track 2 should have ParentComposer1, got %s", tracks[1].Artists[0].Name)
	}
	if tracks[1].Title != "Parent Work 1: Movement 1" {
		t.Errorf("Track 2 title should be 'Parent Work 1: Movement 1', got %q", tracks[1].Title)
	}

	// Track 3: From Parent Work 1
	if tracks[2].Artists[0].Name != "ParentComposer1" {
		t.Errorf("Track 3 should have ParentComposer1, got %s", tracks[2].Artists[0].Name)
	}
	if tracks[2].Title != "Parent Work 1: Movement 2" {
		t.Errorf("Track 3 title should be 'Parent Work 1: Movement 2', got %q", tracks[2].Title)
	}

	// Track 4: From Parent Work 2
	if tracks[3].Artists[0].Name != "ParentComposer2" {
		t.Errorf("Track 4 should have ParentComposer2, got %s", tracks[3].Artists[0].Name)
	}
	if tracks[3].Title != "Parent Work 2: Movement A" {
		t.Errorf("Track 4 title should be 'Parent Work 2: Movement A', got %q", tracks[3].Title)
	}

	// Track 5: Has its own composer
	if tracks[4].Artists[0].Name != "Composer5" {
		t.Errorf("Track 5 should have Composer5, got %s", tracks[4].Artists[0].Name)
	}
}

func TestParseDiscogsPosition(t *testing.T) {

	/*
	   12.2.9. The standard Discogs positions are:

	   Without sides (for example, CD): 1, 2, 3…
	   With sides (for example, LP, 7", cassette): A1, A2…, B1, B2…
	   Multiple 12", LP, etc., just continue the letters: …C1, C2, D1, D2, etc.
	   With programs (for example, 8-track cartridge and 4-track cartridge): A1, A2…, B1, B2…, C1, C2...
	   Multiple CDs, etc.: 1-1, 1-2…, 2-1, 2-2…
	   Multi-disc or multi-format releases: Use a clear and simple position numbering scheme which
	     differentiates each item, for example, CD1-1, DVD1-1, etc.
	   Sub tracks, for example, DJ mixes that comprise one track on a CD: Separate songs or tunes that
	     are rolled into one track on a CD, LP, etc., should be listed using a point and then a number:
	     1, 2, 3.1, 3.2, 3.3, 4,… Letters can also be used, with or without a point: A3.a, A3.b or A3a, A3b,…
	   Enhanced CDs containing extra material, use a prefix in the Track Position field that denotes what the
	     extra material is, for example, "Video 1", "Video 2". Enter the enhanced tracks after the audio material.
	     In the Release Notes, mention any specific software and / or the technology required in order to use the
	     material, for example, "Video material viewable on PC and Mac. Videos launched automatically in a new Window".
	*/

	tests := []struct {
		name      string
		position  string
		wantDisc  int
		wantTrack int
	}{
		{"single number", "1", 1, 1},
		{"disc-track format", "2-10", 2, 10},
		{"CD prefix", "CD3-2", 3, 2},
		{"empty", "", 1, 0},
		{"invalid", "abc", 1, 0},
		{"track 20", "20", 1, 20},
		{"with sides", "B4", 1, 0},       // TODO: should be 2, 1
		{"multiple 12\" LP", "C3", 1, 0}, // TODO: should be 3, 1
		{"with programs", "D8", 1, 0},    // TODO: should be 4, 8
		{"multiple CDs", "3-99", 3, 99},
		{"multi-disc or multi-format releases", "DVD9-88", 9, 88},
		{"sub tracks", "3.4", 1, 0},       // TODO: should be 1, "3.4"
		{"enhanced CDs", "Video 1", 1, 0}, // TODO: should be 1, "Video 1"
		{"enhanced CDs", "Video 2", 1, 0}, // TODO: should be 1, "Video 2"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disc, track := parseDiscogsPosition(tt.position)
			if disc != tt.wantDisc {
				t.Errorf("parseDiscogsPosition(%q) disc = %d, want %d", tt.position, disc, tt.wantDisc)
			}
			if track != tt.wantTrack {
				t.Errorf("parseDiscogsPosition(%q) track = %d, want %d", tt.position, track, tt.wantTrack)
			}
		})
	}
}
func TestConvertDiscogsRelease_ProcessSubtracks(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Tracklist: []Track{
			{Position: "11", Title: "Track 11", Artists: []Artist{{Name: "Composer1", Role: "Composed By"}}},
			{
				Position: "",
				Title:    "Quatre Motets Pour Le Temps de Noël",
				Artists:  []Artist{{Name: "Francis Poulenc", Role: "Composed By"}},
				SubTracks: []Track{
					{Position: "16", Title: "O Magnum Mysterium"},
					{Position: "17", Title: "Quem Vidistis Pastores Dicite"},
					{Position: "18", Title: "Videntes Stellam"},
					{Position: "19", Title: "Hodie Christus Natus Est"},
				},
			},
			{Position: "20", Title: "Track 20", Artists: []Artist{{Name: "Composer2", Role: "Composed By"}}},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}
	if torrent == nil {
		t.Fatal("convertDiscogsRelease returned nil")
	}

	tracks := torrent.Tracks()
	// Should have tracks 11, 16-19 (from subtracks), and 20 = 6 tracks total
	if len(tracks) != 6 {
		t.Fatalf("Expected 6 tracks (11, 16-19 from subtracks, 20), got %d", len(tracks))
	}

	// Check track 11 (before parent work, should be unchanged)
	if tracks[0].Track != 11 {
		t.Errorf("First track should be 11, got %d", tracks[0].Track)
	}
	if tracks[0].Title != "Track 11" {
		t.Errorf("Track 11 title should be unchanged, got %q", tracks[0].Title)
	}

	// Check tracks 16-19 (from subtracks, should have parent work title prepended and parent composer)
	expectedTitles := []string{
		"Quatre Motets Pour Le Temps de Noël: O Magnum Mysterium",
		"Quatre Motets Pour Le Temps de Noël: Quem Vidistis Pastores Dicite",
		"Quatre Motets Pour Le Temps de Noël: Videntes Stellam",
		"Quatre Motets Pour Le Temps de Noël: Hodie Christus Natus Est",
	}

	for i := 1; i <= 4; i++ {
		expectedTrackNum := 15 + i
		track := tracks[i]
		if track.Track != expectedTrackNum {
			t.Errorf("Track %d should be track number %d, got %d", i+1, expectedTrackNum, track.Track)
		}
		expectedTitle := expectedTitles[i-1]
		if track.Title != expectedTitle {
			t.Errorf("Track %d title should be %q, got %q", expectedTrackNum, expectedTitle, track.Title)
		}
		if len(track.Artists) == 0 {
			t.Errorf("Track %d should have composer Francis Poulenc, got no artists", expectedTrackNum)
		} else if track.Artists[0].Name != "Francis Poulenc" {
			t.Errorf("Track %d should have composer Francis Poulenc, got %s", expectedTrackNum, track.Artists[0].Name)
		}
		if track.Artists[0].Role != domain.RoleComposer {
			t.Errorf("Track %d composer should have RoleComposer, got %v", expectedTrackNum, track.Artists[0].Role)
		}
	}

	// Check track 20 (after parent work, should be unchanged)
	if tracks[5].Track != 20 {
		t.Errorf("Last track should be 20, got %d", tracks[5].Track)
	}
	if tracks[5].Title != "Track 20" {
		t.Errorf("Track 20 title should be unchanged, got %q", tracks[5].Title)
	}
}

func TestConvertDiscogsRelease_AlbumArtistsInTracks(t *testing.T) {
	release := &Release{
		Title: "Test Album",
		Year:  2013,
		Artists: []Artist{
			{Name: "RIAS-Kammerchor"},
			{Name: "Hans-Christoph Rademann"},
		},
		ExtraArtists: []Artist{
			{Name: "RIAS-Kammerchor", Role: "Choir"},
			{Name: "Hans-Christoph Rademann", Role: "Chorus Master"},
		},
		Tracklist: []Track{
			{
				Position: "1",
				Title:    "Track 1",
				Artists:  []Artist{{Name: "Composer", Role: "Composed By"}},
			},
			{
				Position: "2",
				Title:    "Track 2",
				// No track-specific artists
			},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}
	if torrent == nil {
		t.Fatal("convertDiscogsRelease returned nil")
	}

	// Check that album artists are included in all tracks
	tracks := torrent.Tracks()
	if len(tracks) != 2 {
		t.Fatalf("Expected 2 tracks, got %d", len(tracks))
	}

	// Track 1 should have composer + album artists
	track1 := tracks[0]
	hasComposer := false
	hasRIAS := false
	hasRademann := false
	for _, artist := range track1.Artists {
		if artist.Name == "Composer" && artist.Role == domain.RoleComposer {
			hasComposer = true
		}
		if artist.Name == "RIAS-Kammerchor" && artist.Role == domain.RoleEnsemble {
			hasRIAS = true
		}
		if artist.Name == "Hans-Christoph Rademann" && artist.Role == domain.RoleConductor {
			hasRademann = true
		}
	}
	if !hasComposer {
		t.Error("Track 1 should have composer")
	}
	if !hasRIAS {
		t.Error("Track 1 should have RIAS-Kammerchor as album artist")
	}
	if !hasRademann {
		t.Error("Track 1 should have Hans-Christoph Rademann as album artist")
	}

	// Track 2 should have album artists even though it has no track-specific artists
	track2 := tracks[1]
	hasRIAS = false
	hasRademann = false
	for _, artist := range track2.Artists {
		if artist.Name == "RIAS-Kammerchor" && artist.Role == domain.RoleEnsemble {
			hasRIAS = true
		}
		if artist.Name == "Hans-Christoph Rademann" && artist.Role == domain.RoleConductor {
			hasRademann = true
		}
	}
	if !hasRIAS {
		t.Error("Track 2 should have RIAS-Kammerchor as album artist")
	}
	if !hasRademann {
		t.Error("Track 2 should have Hans-Christoph Rademann as album artist")
	}
}

func TestConvertDiscogsRelease_RootPathGenerated(t *testing.T) {
	release := &Release{
		Title: "Goldberg Variations",
		Year:  1981,
		Tracklist: []Track{
			{
				Position: "1",
				Title:    "Aria",
				Artists:  []Artist{{Name: "Johann Sebastian Bach", Role: "Composed By"}},
			},
			{
				Position: "2",
				Title:    "Variation 1",
				Artists:  []Artist{{Name: "Johann Sebastian Bach", Role: "Composed By"}},
			},
		},
	}

	torrent, err := release.DomainTorrent("test-path", nil)
	if err != nil {
		t.Fatalf("DomainTorrent() error = %v", err)
	}
	if torrent == nil {
		t.Fatal("convertDiscogsRelease returned nil")
	}

	// RootPath should be generated using GenerateDirectoryName logic
	// Should include composer, title, and year
	rootPath := torrent.RootPath
	if rootPath == "" {
		t.Error("RootPath should not be empty")
	}
	if rootPath == "test-path" {
		t.Error("RootPath should be generated, not use the input parameter")
	}
	// Should contain year
	if !contains(rootPath, "1981") {
		t.Errorf("RootPath should contain year 1981, got %q", rootPath)
	}
	// Should contain composer (last name)
	if !contains(rootPath, "Bach") {
		t.Errorf("RootPath should contain composer Bach, got %q", rootPath)
	}
	// Should contain title
	if !contains(rootPath, "Goldberg") {
		t.Errorf("RootPath should contain title Goldberg, got %q", rootPath)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestMapDiscogsRoleToDomain(t *testing.T) {
	tests := []struct {
		name   string
		artist Artist
		want   domain.Role
	}{
		{"composed by", Artist{Name: "Bach", Role: "Composed By"}, domain.RoleComposer},
		{"composer", Artist{Name: "Bach", Role: "Composer"}, domain.RoleComposer},
		{"conductor", Artist{Name: "Karajan", Role: "Conductor"}, domain.RoleConductor},
		{"choir", Artist{Name: "Choir", Role: "Choir"}, domain.RoleEnsemble},
		{"chorus master", Artist{Name: "Master", Role: "Chorus Master"}, domain.RoleConductor},
		{"orchestra", Artist{Name: "Orchestra", Role: "Orchestra"}, domain.RoleEnsemble},
		{"empty role with ensemble name", Artist{Name: "Berlin Philharmonic", Role: ""}, domain.RoleEnsemble},
		{"empty role with soloist name", Artist{Name: "Vladimir Horowitz", Role: ""}, domain.RoleUnknown},
		{"unknown role", Artist{Name: "Artist", Role: "Unknown"}, domain.RoleUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.artist.DomainRole(nil, nil)
			if got != tt.want {
				t.Errorf("Artist{Name: %q, Role: %q}.DomainRole() = %v, want %v", tt.artist.Name, tt.artist.Role, got, tt.want)
			}
		})
	}
}
