package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestRules_Discover tests that rule discovery finds valid rule methods
func TestRules_TorrentRules(t *testing.T) {
	rules := NewRules()
	torrentRules := rules.TorrentRules()

	// Should discover at least one rule
	if len(torrentRules) == 0 {
		t.Fatal("Discover() found no rules")
	}

	t.Logf("Discovered %d rules", len(torrentRules))

	// Each discovered rule should be callable
	emptyTorrent := &domain.Torrent{Title: "Test", OriginalYear: 2000}
	for i, ruleFunc := range torrentRules {
		result := ruleFunc(emptyTorrent, nil)

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
	emptyTrack := &domain.Track{Disc: 1, Track: 1, Title: "Test", File: domain.File{Path: "01 Test.flac"}}
	emptyTorrent := &domain.Torrent{Title: "Test", OriginalYear: 2000}
	for i, ruleFunc := range trackRules {
		result := ruleFunc(emptyTrack, nil, emptyTorrent, nil)
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
	albumRules := rules.TorrentRules()

	emptyTorrent := &domain.Torrent{Title: "Test", OriginalYear: 2000}
	seenIDs := make(map[string]int)

	for i, ruleFunc := range albumRules {
		result := ruleFunc(emptyTorrent, nil)
		ruleID := result.Meta.ID

		if prevIndex, exists := seenIDs[ruleID]; exists {
			t.Errorf("Duplicate rule ID %q: found at indices %d and %d", ruleID, prevIndex, i)
		}
		seenIDs[ruleID] = i
	}

	emptyTrack := &domain.Track{Disc: 1, Track: 1, Title: "Test", File: domain.File{Path: "01 Test.flac"}}
	trackRules := rules.TrackRules()
	for i, ruleFunc := range trackRules {
		result := ruleFunc(emptyTrack, nil, emptyTorrent, nil)
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
	torrent := &domain.Torrent{Title: "Empty Album", OriginalYear: 2000}
	ruleCount := 0
	errorCount := 0
	mapAllRulesWithTorrent(torrent, func(result RuleResult) {
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
	torrent := &domain.Torrent{
		Title: "Test Album", OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "test label",
			Year:          2013,
			CatalogNumber: "HMC902170",
		},
		Files: []domain.FileLike{
			&domain.Track{
				File:  domain.File{Path: "01 Test Work.flac"},
				Disc:  1,
				Track: 1,
				Title: "Test Work",
			},
		},
	}

	// Use same torrent as reference
	ruleCount := 0
	errorCount := 0
	warningCount := 0
	tracks := torrent.Tracks()
	if len(tracks) == 0 {
		t.Fatal("No tracks found in torrent")
	}
	track := tracks[0]
	mapAllRulesWithReferenceTorrentAndTrack(track, track, torrent, torrent, func(result RuleResult) {
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
	torrent := &domain.Torrent{Title: "Test", OriginalYear: 2000}
	mapAllRulesWithTorrent(torrent, fn)
}

func mapAllRulesWithTorrent(torrent *domain.Torrent, fn func(result RuleResult)) {
	track := &domain.Track{Disc: 1, Track: 1, Title: "Test", File: domain.File{Path: "01 Test.flac"}}
	mapAllRulesWithTorrentAndTrack(track, torrent, fn)
}

func mapAllRulesWithTorrentAndTrack(track *domain.Track, torrent *domain.Torrent, fn func(result RuleResult)) {
	mapAllRulesWithReferenceTorrentAndTrack(track, nil, torrent, nil, fn)
}

func mapAllRulesWithReferenceTorrentAndTrack(track *domain.Track, referenceTrack *domain.Track, torrent *domain.Torrent, referenceTorrent *domain.Torrent, fn func(result RuleResult)) {
	rules := NewRules()
	albumRules := rules.TorrentRules()

	for _, ruleFunc := range albumRules {
		fn(ruleFunc(torrent, referenceTorrent))
	}

	if track != nil {
		trackRules := rules.TrackRules()
		for _, ruleFunc := range trackRules {
			fn(ruleFunc(track, referenceTrack, torrent, referenceTorrent))
		}
	}
}
