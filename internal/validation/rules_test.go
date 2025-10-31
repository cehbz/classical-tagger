package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestRules_Discover tests that rule discovery finds valid rule methods
func TestRules_AlbumRules(t *testing.T) {
	rules := NewRules()
	albumRules := rules.AlbumRules()

	// Should discover at least one rule
	if len(albumRules) == 0 {
		t.Fatal("Discover() found no rules")
	}

	t.Logf("Discovered %d rules", len(albumRules))

	// Each discovered rule should be callable
	emptyAlbum := &domain.Album{Title: "Test", OriginalYear: 2000}
	for i, ruleFunc := range albumRules {
		result := ruleFunc(emptyAlbum, nil)

		// Should return valid RuleResult
		if result.Meta.ID == "" {
			t.Errorf("Rule %d has empty ID", i)
		}
		if result.Meta.Name == "" {
			t.Errorf("Rule %d (%s) has empty name", i, result.Meta.ID)
		}
		if result.Meta.Level.String() == "" || result.Meta.Level.String() == "UNKNOWN" {
			t.Errorf("Rule %d (%s) has invalid level: %d", i, result.Meta.ID, result.Meta.Level)
		}
	}
}

// TestRules_TrackRules tests that rule discovery finds valid track rules
func TestRules_TrackRules(t *testing.T) {
	rules := NewRules()
	trackRules := rules.TrackRules()

	if len(trackRules) == 0 {
		t.Fatal("TrackRules() found no rules")
	}

	t.Logf("Discovered %d rules", len(trackRules))

	// Each discovered rule should be callable
	emptyTrack := &domain.Track{Disc: 1, Track: 1, Title: "Test", Name: "01 Test.flac"}
	emptyAlbum := &domain.Album{Title: "Test", OriginalYear: 2000}
	for i, ruleFunc := range trackRules {
		result := ruleFunc(emptyTrack, nil, emptyAlbum, nil)
		if result.Meta.ID == "" {
			t.Errorf("Rule %d has empty ID", i)
		}
		if result.Meta.Name == "" {
			t.Errorf("Rule %d (%s) has empty name", i, result.Meta.ID)
		}
		if result.Meta.Weight <= 0 {
			t.Errorf("Rule %d (%s) has invalid Weight: %f", i, result.Meta.ID, result.Meta.Weight)
		}
	}
}

// TestRules_DiscoverUniqueIDs tests that all discovered rules have unique IDs
func TestRules_DiscoverUniqueIDs(t *testing.T) {
	rules := NewRules()
	albumRules := rules.AlbumRules()

	emptyAlbum := &domain.Album{Title: "Test", OriginalYear: 2000}
	seenIDs := make(map[string]int)

	for i, ruleFunc := range albumRules {
		result := ruleFunc(emptyAlbum, nil)
		ruleID := result.Meta.ID

		if prevIndex, exists := seenIDs[ruleID]; exists {
			t.Errorf("Duplicate rule ID %q: found at indices %d and %d", ruleID, prevIndex, i)
		}
		seenIDs[ruleID] = i
	}

	emptyTrack := &domain.Track{Disc: 1, Track: 1, Title: "Test", Name: "01 Test.flac"}
	trackRules := rules.TrackRules()
	for i, ruleFunc := range trackRules {
		result := ruleFunc(emptyTrack, nil, emptyAlbum, nil)
		ruleID := result.Meta.ID
		if prevIndex, exists := seenIDs[ruleID]; exists {
			t.Errorf("Duplicate rule ID %q: found at indices %d and %d", ruleID, prevIndex, i)
		}
		seenIDs[ruleID] = i
	}

	t.Logf("Verified %d unique rule IDs", len(seenIDs))
}

// TestRules_DiscoverIgnoresDiscoverMethod tests that Discover() doesn't discover itself
func TestRules_DiscoverIgnoresDiscoverMethod(t *testing.T) {
	mapAllRules(func(result RuleResult) {
		if result.Meta.Name == "Discover" {
			t.Error("Discover() incorrectly discovered itself as a rule")
		}
	})
}

// TestRunAll_EmptyAlbum tests running all rules on an empty album
func TestRunAll_EmptyAlbum(t *testing.T) {
	album := &domain.Album{Title: "Empty Album", OriginalYear: 2000}
	ruleCount := 0
	errorCount := 0
	mapAllRulesWithAlbum(album, func(result RuleResult) {
		ruleCount++
		if !result.Passed() {
			errorCount++
		}
	})

	if ruleCount == 0 {
		t.Fatal("No rules were discovered")
	}
	if errorCount == 0 {
		t.Errorf("Expected at least one error, got %d", errorCount)
	}
}

// TestRunAll_ValidAlbum tests running all rules on a valid album
func TestRunAll_ValidAlbum(t *testing.T) {
	// Create a valid album
	album := &domain.Album{
		Title: "Test Album", OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "test label",
			Year:          2013,
			CatalogNumber: "HMC902170",
		},
		Tracks: []*domain.Track{
			&domain.Track{
				Disc:  1,
				Track: 1,
				Title: "Test Work",
				Artists: []domain.Artist{
					{Name: "Felix Mendelssohn", Role: domain.RoleComposer},
					{Name: "Test Ensemble", Role: domain.RoleEnsemble},
				},
				Name: "01 Test Work.flac",
			},
		},
	}

	// Use same album as reference
	ruleCount := 0
	errorCount := 0
	warningCount := 0
	mapAllRulesWithReferenceAlbumAndTrack(album.Tracks[0], album.Tracks[0], album, album, func(result RuleResult) {
		ruleCount++
		if !result.Passed() {
			errorCount++
		}
		for _, issue := range result.Issues {
			if issue.Level == domain.LevelError {
				errorCount++
			}
		}
	})

	// Should have some passed rules
	if ruleCount == errorCount {
		t.Error("Expected some rules to pass for valid album")
	}

	t.Logf("Results: %d/%d passed, %d errors, %d warnings",
		ruleCount-errorCount-warningCount, ruleCount, errorCount, warningCount)
}

func mapAllRules(fn func(result RuleResult)) {
	album := &domain.Album{Title: "Test", OriginalYear: 2000}
	mapAllRulesWithAlbum(album, fn)
}

func mapAllRulesWithAlbum(album *domain.Album, fn func(result RuleResult)) {
	track := &domain.Track{Disc: 1, Track: 1, Title: "Test", Name: "01 Test.flac"}
	mapAllRulesWithAlbumAndTrack(track, album, fn)
}

func mapAllRulesWithAlbumAndTrack(track *domain.Track, album *domain.Album, fn func(result RuleResult)) {
	mapAllRulesWithReferenceAlbumAndTrack(track, nil, album, nil, fn)
}

func mapAllRulesWithReferenceAlbumAndTrack(track, referenceTrack *domain.Track, album, referenceAlbum *domain.Album, fn func(result RuleResult)) {
	rules := NewRules()
	albumRules := rules.AlbumRules()

	for _, ruleFunc := range albumRules {
		fn(ruleFunc(album, referenceAlbum))
	}

	if track != nil {
		trackRules := rules.TrackRules()
		for _, ruleFunc := range trackRules {
			fn(ruleFunc(track, referenceTrack, album, referenceAlbum))
		}
	}
}
