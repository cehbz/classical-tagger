package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestRules_Discover tests that rule discovery finds valid rule methods
func TestRules_Discover(t *testing.T) {
	rules := NewRules()
	discovered := rules.Discover()

	// Should discover at least one rule
	if len(discovered) == 0 {
		t.Fatal("Discover() found no rules")
	}

	t.Logf("Discovered %d rules", len(discovered))

	// Each discovered rule should be callable
	empty := &domain.Album{Title: "Test", OriginalYear: 2000}
	for i, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)

		// Should return valid RuleResult
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
	discovered := rules.Discover()

	empty := &domain.Album{Title: "Test", OriginalYear: 2000}
	seenIDs := make(map[string]int)

	for i, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)
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
	rules := NewRules()
	discovered := rules.Discover()

	empty := &domain.Album{Title: "Test", OriginalYear: 2000}

	for _, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)
		if result.Meta.Name == "Discover" {
			t.Error("Discover() incorrectly discovered itself as a rule")
		}
	}
}

// TestAll_ConvenienceFunction tests that All() works as expected
func TestAll_ConvenienceFunction(t *testing.T) {
	allRules := All()

	if len(allRules) == 0 {
		t.Fatal("All() returned no rules")
	}

	// Should be the same as NewRules().Discover()
	rules := NewRules()
	discovered := rules.Discover()

	if len(allRules) != len(discovered) {
		t.Errorf("All() count = %d, Discover() count = %d, should match", len(allRules), len(discovered))
	}
}

// TestRunAll_EmptyAlbum tests running all rules on an empty album
func TestRunAll_EmptyAlbum(t *testing.T) {
	album := &domain.Album{Title: "Empty Album", OriginalYear: 2000}
	reference := &domain.Album{Title: "Empty Reference", OriginalYear: 2000}

	rules := All()
	result := RunAll(album, reference, rules)

	if result == nil {
		t.Fatal("RunAll() returned nil")
	}

	// Should have counted all rules
	if result.TotalCount() != len(rules) {
		t.Errorf("TotalRules() = %d, want %d", result.TotalCount(), len(rules))
	}

	// Album with no tracks should have at least one error
	if result.ErrorCount() == 0 {
		t.Error("Expected at least one error for album with no tracks")
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
	reference := album

	rules := All()
	result := RunAll(album, reference, rules)

	// Should have some passed rules
	if result.PassedCount() == 0 {
		t.Error("Expected some rules to pass for valid album")
	}

	t.Logf("Results: %d/%d passed, %d errors, %d warnings",
		result.PassedCount(), result.TotalCount(), result.ErrorCount(), result.WarningCount())
}

// TestValidationResult_ImprovementScore tests score calculation
func TestValidationResult_ImprovementScore(t *testing.T) {
	tests := []struct {
		Name           string
		PassedRules    int
		FailedRules    int
		ExpectedScore  float64
		ScoreThreshold float64 // For floating point comparison
	}{
		{
			Name:           "all pass",
			PassedRules:    10,
			FailedRules:    0,
			ExpectedScore:  1.0,
			ScoreThreshold: 0.001,
		},
		{
			Name:           "all fail",
			PassedRules:    0,
			FailedRules:    10,
			ExpectedScore:  0.0,
			ScoreThreshold: 0.001,
		},
		{
			Name:           "half pass",
			PassedRules:    5,
			FailedRules:    5,
			ExpectedScore:  0.5,
			ScoreThreshold: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := &ValidationResult{
				totalRules:  tt.PassedRules + tt.FailedRules,
				passedRules: tt.PassedRules,
				failedRules: tt.FailedRules,
				ruleResults: make([]RuleResult, tt.PassedRules+tt.FailedRules),
			}

			// Create mock rule results (all weight 1.0)
			for i := 0; i < tt.PassedRules; i++ {
				meta := RuleMetadata{ID: "pass", Name: "pass", Level: domain.LevelError, Weight: 1.0}
				result.ruleResults[i] = RuleResult{Meta: meta}
			}
			for i := 0; i < tt.FailedRules; i++ {
				meta := RuleMetadata{ID: "fail", Name: "fail", Level: domain.LevelError, Weight: 1.0}
				issue := domain.ValidationIssue{
					Level:   domain.LevelError,
					Track:   1,
					Rule:    "fail",
					Message: "failed",
				}
				result.ruleResults[tt.PassedRules+i] = RuleResult{Meta: meta, Issues: []domain.ValidationIssue{issue}}
			}

			score := result.ImprovementScore()

			diff := score - tt.ExpectedScore
			if diff < -tt.ScoreThreshold || diff > tt.ScoreThreshold {
				t.Errorf("ImprovementScore() = %f, want %f (±%f)", score, tt.ExpectedScore, tt.ScoreThreshold)
			}
		})
	}
}

// TestValidationResult_FailedRules tests filtering for failed rules
func TestValidationResult_FailedRules(t *testing.T) {
	album := &domain.Album{
		Title: "Test Album", OriginalYear: 2013,
		Tracks: []*domain.Track{
			// Create track with composer in title (will fail rule)
			&domain.Track{
				Disc:    1,
				Track:   1,
				Title:   "Bach: Goldberg Variations",
				Artists: []domain.Artist{{Name: "Johann Sebastian Bach", Role: domain.RoleComposer}},
				Name:    "01 Bach: Goldberg Variations.flac",
			},
		},
	}

	reference := album
	rules := All()
	result := RunAll(album, reference, rules)

	failedRules := result.FailedRules()

	// Should have at least the composer-in-title failure
	if len(failedRules) == 0 {
		t.Error("Expected some failed rules for album with violations")
	}

	// All failed rules should have issues
	for _, rr := range failedRules {
		if rr.Passed() {
			t.Errorf("FailedRules() included passing rule %q", rr.Meta.ID)
		}
		if len(rr.Issues) == 0 {
			t.Errorf("Failed rule %q has no issues", rr.Meta.ID)
		}
	}
}

// TestValidationResult_RuleByID tests finding specific rules
func TestValidationResult_RuleByID(t *testing.T) {
	album := &domain.Album{Title: "Test Album", OriginalYear: 2013}
	reference := album

	rules := All()
	result := RunAll(album, reference, rules)

	// Get all rule IDs
	empty := &domain.Album{Title: "Test", OriginalYear: 2000}
	ruleIDs := make([]string, 0)
	for _, ruleFunc := range rules {
		rr := ruleFunc(empty, empty)
		ruleIDs = append(ruleIDs, rr.Meta.ID)
	}

	// Should be able to find each rule
	for _, id := range ruleIDs {
		found := result.RuleByID(id)
		if found == nil {
			t.Errorf("RuleByID(%q) returned nil", id)
		} else if found.Meta.ID != id {
			t.Errorf("RuleByID(%q) returned rule with ID %q", id, found.Meta.ID)
		}
	}

	// Non-existent ID should return nil
	notFound := result.RuleByID("nonexistent.rule")
	if notFound != nil {
		t.Error("RuleByID() should return nil for non-existent ID")
	}
}

// TestValidationResult_Counters tests various counter methods
func TestValidationResult_Counters(t *testing.T) {
	album := &domain.Album{Title: "Test Album", OriginalYear: 2013}
	reference := album

	rules := All()
	result := RunAll(album, reference, rules)

	// TotalRules should equal number of rules
	if result.TotalCount() != len(rules) {
		t.Errorf("TotalRules() = %d, want %d", result.TotalCount(), len(rules))
	}

	// PassedRules + FailedRules should equal TotalRules
	if result.PassedCount()+result.FailedCount() != result.TotalCount() {
		t.Errorf("PassedRules(%d) + FailedRules(%d) != TotalRules(%d)",
			result.PassedCount(), result.FailedCount(), result.TotalCount())
	}

	// Error and warning counts should be non-negative
	if result.ErrorCount() < 0 {
		t.Errorf("ErrorCount() = %d, should be non-negative", result.ErrorCount())
	}
	if result.WarningCount() < 0 {
		t.Errorf("WarningCount() = %d, should be non-negative", result.WarningCount())
	}
}

// TestRules_MethodSignatureValidation tests that only methods with correct signature are discovered
func TestRules_MethodSignatureValidation(t *testing.T) {
	// This is implicitly tested by Discover(), but we make it explicit
	// Discover() should only find methods matching:
	// func (r *Rules) MethodName(actual, reference *domain.Album) RuleResult

	rules := NewRules()
	discovered := rules.Discover()

	// Call each discovered rule to ensure it matches signature
	album := &domain.Album{Title: "Test", OriginalYear: 2000}

	for i, ruleFunc := range discovered {
		// Should not panic when called
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Rule %d panicked when called: %v", i, r)
			}
		}()

		result := ruleFunc(album, album)

		// Should return valid RuleResult
		_ = result.Meta
		_ = result.Issues
		_ = result.Passed()
	}
}

// TestRules_CoverageChecklist tests that expected rules are discovered
func TestRules_CoverageChecklist(t *testing.T) {
	// This test documents expected rules - update as rules are added
	expectedRules := map[string]bool{
		// Will be implemented:
		"2.3.12":             false, // Path length
		"2.3.14":             false, // Track sort order
		"2.3.20":             false, // No leading spaces
		"classical.composer": false, // Composer not in title
		// Add more as implemented...
	}

	rules := NewRules()
	discovered := rules.Discover()

	empty := &domain.Album{Title: "Test", OriginalYear: 2000}
	foundRules := make(map[string]bool)

	for _, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)
		foundRules[result.Meta.ID] = true
	}

	// Report on expected rules
	for ruleID, shouldExist := range expectedRules {
		exists := foundRules[ruleID]
		if shouldExist && !exists {
			t.Errorf("Expected rule %q not found", ruleID)
		}
		if exists {
			t.Logf("✓ Found rule %q", ruleID)
		}
	}

	// This test will evolve as we add rules
	t.Logf("Total rules discovered: %d", len(discovered))
}
