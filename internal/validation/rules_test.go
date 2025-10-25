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
	empty, _ := domain.NewAlbum("Test", 2000)
	for i, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)
		
		// Should return valid RuleResult
		if result.Meta().ID() == "" {
			t.Errorf("Rule %d has empty ID", i)
		}
		if result.Meta().Name() == "" {
			t.Errorf("Rule %d (%s) has empty name", i, result.Meta().ID())
		}
		if result.Meta().Weight() <= 0 {
			t.Errorf("Rule %d (%s) has invalid weight: %f", i, result.Meta().ID(), result.Meta().Weight())
		}
	}
}

// TestRules_DiscoverUniqueIDs tests that all discovered rules have unique IDs
func TestRules_DiscoverUniqueIDs(t *testing.T) {
	rules := NewRules()
	discovered := rules.Discover()
	
	empty, _ := domain.NewAlbum("Test", 2000)
	seenIDs := make(map[string]int)
	
	for i, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)
		ruleID := result.Meta().ID()
		
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
	
	empty, _ := domain.NewAlbum("Test", 2000)
	
	for _, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)
		if result.Meta().Name() == "Discover" {
			t.Error("Discover() incorrectly discovered itself as a rule")
		}
	}
}

// TestRuleMetadata_PassFail tests Pass() and Fail() helper methods
func TestRuleMetadata_PassFail(t *testing.T) {
	meta := RuleMetadata{
		id:     "test.rule",
		name:   "Test Rule",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	// Test Pass()
	passResult := meta.Pass()
	if !passResult.Passed() {
		t.Error("Pass() should create passing result")
	}
	if len(passResult.Issues()) != 0 {
		t.Errorf("Pass() should have no issues, got %d", len(passResult.Issues()))
	}
	if passResult.Meta().ID() != meta.id {
		t.Errorf("Pass() meta.ID = %q, want %q", passResult.Meta().ID(), meta.id)
	}
	
	// Test Fail()
	issue1 := domain.NewIssue(domain.LevelError, 1, "test.rule", "Issue 1")
	issue2 := domain.NewIssue(domain.LevelError, 2, "test.rule", "Issue 2")
	
	failResult := meta.Fail(issue1, issue2)
	if failResult.Passed() {
		t.Error("Fail() should create failing result")
	}
	if len(failResult.Issues()) != 2 {
		t.Errorf("Fail() issues = %d, want 2", len(failResult.Issues()))
	}
	if failResult.Meta().ID() != meta.id {
		t.Errorf("Fail() meta.ID = %q, want %q", failResult.Meta().ID(), meta.id)
	}
}

// TestRuleResult_Getters tests that RuleResult getters work correctly
func TestRuleResult_Getters(t *testing.T) {
	meta := RuleMetadata{
		id:     "2.3.12",
		name:   "Path length check",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	issue := domain.NewIssue(domain.LevelError, 1, "2.3.12", "Path too long")
	result := meta.Fail(issue)
	
	// Test Meta()
	resultMeta := result.Meta()
	if resultMeta.ID() != "2.3.12" {
		t.Errorf("Meta().ID() = %q, want %q", resultMeta.ID(), "2.3.12")
	}
	if resultMeta.Name() != "Path length check" {
		t.Errorf("Meta().Name() = %q, want %q", resultMeta.Name(), "Path length check")
	}
	if resultMeta.Level() != domain.LevelError {
		t.Errorf("Meta().Level() = %v, want %v", resultMeta.Level(), domain.LevelError)
	}
	if resultMeta.Weight() != 1.0 {
		t.Errorf("Meta().Weight() = %f, want %f", resultMeta.Weight(), 1.0)
	}
	
	// Test Issues()
	issues := result.Issues()
	if len(issues) != 1 {
		t.Fatalf("Issues() count = %d, want 1", len(issues))
	}
	if issues[0].Message() != "Path too long" {
		t.Errorf("Issues()[0].Message() = %q, want %q", issues[0].Message(), "Path too long")
	}
	
	// Test Passed()
	if result.Passed() {
		t.Error("Result with issues should not pass")
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
	album, _ := domain.NewAlbum("Empty Album", 2000)
	reference, _ := domain.NewAlbum("Empty Reference", 2000)
	
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
	album, _ := domain.NewAlbum("Test Album", 2013)
	edition, _ := domain.NewEdition("test label", 2013)
	edition = edition.WithCatalogNumber("HMC902170")
	album = album.WithEdition(edition)
	
	composer, _ := domain.NewArtist("Felix Mendelssohn", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Test Ensemble", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Test Work", []domain.Artist{composer, ensemble})
	track = track.WithName("01 Test Work.flac")
	album.AddTrack(track)
	
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
		name           string
		passedRules    int
		failedRules    int
		expectedScore  float64
		scoreThreshold float64 // For floating point comparison
	}{
		{
			name:           "all pass",
			passedRules:    10,
			failedRules:    0,
			expectedScore:  1.0,
			scoreThreshold: 0.001,
		},
		{
			name:           "all fail",
			passedRules:    0,
			failedRules:    10,
			expectedScore:  0.0,
			scoreThreshold: 0.001,
		},
		{
			name:           "half pass",
			passedRules:    5,
			failedRules:    5,
			expectedScore:  0.5,
			scoreThreshold: 0.001,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				totalRules:  tt.passedRules + tt.failedRules,
				passedRules: tt.passedRules,
				failedRules: tt.failedRules,
				ruleResults: make([]RuleResult, tt.passedRules+tt.failedRules),
			}
			
			// Create mock rule results (all weight 1.0)
			for i := 0; i < tt.passedRules; i++ {
				meta := RuleMetadata{id: "pass", name: "pass", level: domain.LevelError, weight: 1.0}
				result.ruleResults[i] = meta.Pass()
			}
			for i := 0; i < tt.failedRules; i++ {
				meta := RuleMetadata{id: "fail", name: "fail", level: domain.LevelError, weight: 1.0}
				issue := domain.NewIssue(domain.LevelError, 1, "fail", "failed")
				result.ruleResults[tt.passedRules+i] = meta.Fail(issue)
			}
			
			score := result.ImprovementScore()
			
			diff := score - tt.expectedScore
			if diff < -tt.scoreThreshold || diff > tt.scoreThreshold {
				t.Errorf("ImprovementScore() = %f, want %f (±%f)", score, tt.expectedScore, tt.scoreThreshold)
			}
		})
	}
}

// TestValidationResult_FailedRules tests filtering for failed rules
func TestValidationResult_FailedRules(t *testing.T) {
	album, _ := domain.NewAlbum("Test Album", 2013)
	composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	
	// Create track with composer in title (will fail rule)
	badTrack, _ := domain.NewTrack(1, 1, "Bach: Goldberg Variations", []domain.Artist{composer})
	album.AddTrack(badTrack)
	
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
			t.Errorf("FailedRules() included passing rule %q", rr.Meta().ID())
		}
		if len(rr.Issues()) == 0 {
			t.Errorf("Failed rule %q has no issues", rr.Meta().ID())
		}
	}
}

// TestValidationResult_RuleByID tests finding specific rules
func TestValidationResult_RuleByID(t *testing.T) {
	album, _ := domain.NewAlbum("Test Album", 2013)
	reference := album
	
	rules := All()
	result := RunAll(album, reference, rules)
	
	// Get all rule IDs
	empty, _ := domain.NewAlbum("Test", 2000)
	ruleIDs := make([]string, 0)
	for _, ruleFunc := range rules {
		rr := ruleFunc(empty, empty)
		ruleIDs = append(ruleIDs, rr.Meta().ID())
	}
	
	// Should be able to find each rule
	for _, id := range ruleIDs {
		found := result.RuleByID(id)
		if found == nil {
			t.Errorf("RuleByID(%q) returned nil", id)
		} else if found.Meta().ID() != id {
			t.Errorf("RuleByID(%q) returned rule with ID %q", id, found.Meta().ID())
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
	album, _ := domain.NewAlbum("Test Album", 2013)
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
	album, _ := domain.NewAlbum("Test", 2000)
	
	for i, ruleFunc := range discovered {
		// Should not panic when called
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Rule %d panicked when called: %v", i, r)
			}
		}()
		
		result := ruleFunc(album, album)
		
		// Should return valid RuleResult
		_ = result.Meta()
		_ = result.Issues()
		_ = result.Passed()
	}
}

// TestRules_CoverageChecklist tests that expected rules are discovered
func TestRules_CoverageChecklist(t *testing.T) {
	// This test documents expected rules - update as rules are added
	expectedRules := map[string]bool{
		// Will be implemented:
		"2.3.12":               false, // Path length
		"2.3.14":               false, // Track sort order
		"2.3.20":               false, // No leading spaces
		"classical.composer":   false, // Composer not in title
		// Add more as implemented...
	}
	
	rules := NewRules()
	discovered := rules.Discover()
	
	empty, _ := domain.NewAlbum("Test", 2000)
	foundRules := make(map[string]bool)
	
	for _, ruleFunc := range discovered {
		result := ruleFunc(empty, empty)
		foundRules[result.Meta().ID()] = true
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