package validation

import (
	"github.com/cehbz/classical-tagger/internal/domain"
)

// RuleMetadata describes a rule's identity and severity
type RuleMetadata struct {
	id     string
	name   string
	level  domain.Level
	weight float64
}

// ID returns the rule section identifier (e.g., "2.3.12", "classical.composer")
func (m RuleMetadata) ID() string {
	return m.id
}

// Name returns the human-readable rule name
func (m RuleMetadata) Name() string {
	return m.name
}

// Level returns the severity level (LevelError or LevelWarning)
func (m RuleMetadata) Level() domain.Level {
	return m.level
}

// Weight returns the rule's weight for scoring (typically 1.0)
func (m RuleMetadata) Weight() float64 {
	return m.weight
}

// Pass returns a passing result (no issues)
func (m RuleMetadata) Pass() RuleResult {
	return RuleResult{
		meta:   m,
		issues: nil,
	}
}

// Fail returns a failing result with the given issues
func (m RuleMetadata) Fail(issues ...domain.ValidationIssue) RuleResult {
	return RuleResult{
		meta:   m,
		issues: issues,
	}
}

// RuleResult is what every rule method returns
type RuleResult struct {
	meta   RuleMetadata
	issues []domain.ValidationIssue
}

// Meta returns the rule's metadata
func (r RuleResult) Meta() RuleMetadata {
	return r.meta
}

// Issues returns the validation issues found by this rule
func (r RuleResult) Issues() []domain.ValidationIssue {
	if r.issues == nil {
		return []domain.ValidationIssue{}
	}
	return r.issues
}

// Passed returns true if the rule found no issues
func (r RuleResult) Passed() bool {
	return len(r.issues) == 0
}
