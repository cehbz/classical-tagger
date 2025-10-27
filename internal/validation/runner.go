package validation

import (
	"math"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ValidationResult aggregates results from all validation rules
type ValidationResult struct {
	albumPath    string
	ruleResults  []RuleResult
	totalRules   int
	passedRules  int
	failedRules  int
	errorCount   int
	warningCount int
}

// AlbumPath returns the path to the album being validated
func (r *ValidationResult) AlbumPath() string {
	return r.albumPath
}

// RuleResults returns all individual rule results
func (r *ValidationResult) RuleResults() []RuleResult {
	return r.ruleResults
}

// TotalRules returns the total number of rules executed
func (r *ValidationResult) TotalCount() int {
	return r.totalRules
}

// PassedRules returns the number of rules that passed
func (r *ValidationResult) PassedCount() int {
	return r.passedRules
}

// FailedRules returns the number of rules that failed
func (r *ValidationResult) FailedCount() int {
	return r.failedRules
}

// ErrorCount returns the total number of ERROR level issues
func (r *ValidationResult) ErrorCount() int {
	return r.errorCount
}

// WarningCount returns the total number of WARNING level issues
func (r *ValidationResult) WarningCount() int {
	return r.warningCount
}

// ImprovementScore calculates a score from 0.0 (all rules failed) to 1.0 (all rules passed)
// Each rule contributes its weight to the maximum penalty. Failed rules add penalty.
func (r *ValidationResult) ImprovementScore() float64 {
	maxPenalty := 0.0
	for _, rr := range r.ruleResults {
		maxPenalty += rr.Meta.Weight
	}

	if maxPenalty == 0 {
		return 0
	}

	actualPenalty := 0.0
	for _, rr := range r.ruleResults {
		if !rr.Passed() {
			actualPenalty += rr.Meta.Weight
		}
	}

	return math.Max(0, 1.0-(actualPenalty/maxPenalty))
}

// FailedRules returns only the rules that had issues
func (r *ValidationResult) FailedRules() []RuleResult {
	var failed []RuleResult
	for _, rr := range r.ruleResults {
		if !rr.Passed() {
			failed = append(failed, rr)
		}
	}
	return failed
}

// RuleByID finds a specific rule result by its ID
func (r *ValidationResult) RuleByID(id string) *RuleResult {
	for _, rr := range r.ruleResults {
		if rr.Meta.ID == id {
			return &rr
		}
	}
	return nil
}

// RunAll executes all provided rules and aggregates the results
func RunAll(actual, reference *domain.Album, rules []RuleFunc) *ValidationResult {
	result := &ValidationResult{
		albumPath:   actual.Title, // Use title as path for now
		totalRules:  len(rules),
		ruleResults: make([]RuleResult, 0, len(rules)),
	}

	for _, ruleFunc := range rules {
		ruleResult := ruleFunc(actual, reference)
		result.ruleResults = append(result.ruleResults, ruleResult)

		if ruleResult.Passed() {
			result.passedRules++
		} else {
			result.failedRules++

			// Count errors and warnings
			for _, issue := range ruleResult.Issues {
				switch issue.Level {
				case domain.LevelError:
					result.errorCount++
				case domain.LevelWarning:
					result.warningCount++
				}
			}
		}
	}

	return result
}
