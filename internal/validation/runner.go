package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ValidationResult aggregates results from all validation rules
type ValidationResult struct {
	EntityName      string
	RuleResults     []RuleResult
	TotalRuleCount  int
	PassedRuleCount int
	FailedRuleCount int
	ErrorCount      int
	WarningCount    int
}

// RunAll executes all provided rules and aggregates the results
func RunAllAlbumRules(actual, reference *domain.Album, rules []AlbumRuleFunc) *ValidationResult {
	result := &ValidationResult{
		EntityName:     actual.Title, // Use title as path for now
		TotalRuleCount: len(rules),
		RuleResults:    make([]RuleResult, 0, len(rules)),
	}

	for _, ruleFunc := range rules {
		ruleResult := ruleFunc(actual, reference)
		result.RuleResults = append(result.RuleResults, ruleResult)

		if ruleResult.Passed() {
			result.PassedRuleCount++
		} else {
			result.FailedRuleCount++

			// Count errors and warnings
			for _, issue := range ruleResult.Issues {
				switch issue.Level {
				case domain.LevelError:
					result.ErrorCount++
				case domain.LevelWarning:
					result.WarningCount++
				}
			}
		}
	}

	return result
}

// RunAll executes all provided rules and aggregates the results
func RunAllTrackRules(actualTrack, refTrack *domain.Track, actualAlbum, refAlbum *domain.Album, rules []TrackRuleFunc) *ValidationResult {
	result := &ValidationResult{
		EntityName:     fmt.Sprintf("%s/%s", actualAlbum.Title, actualTrack.Name),
		TotalRuleCount: len(rules),
		RuleResults:    make([]RuleResult, 0, len(rules)),
	}

	for _, ruleFunc := range rules {
		ruleResult := ruleFunc(actualTrack, refTrack, actualAlbum, refAlbum)
		result.RuleResults = append(result.RuleResults, ruleResult)

		if ruleResult.Passed() {
			result.PassedRuleCount++
		} else {
			result.FailedRuleCount++

			// Count errors and warnings
			for _, issue := range ruleResult.Issues {
				switch issue.Level {
				case domain.LevelError:
					result.ErrorCount++
				case domain.LevelWarning:
					result.WarningCount++
				}
			}
		}
	}

	return result
}
