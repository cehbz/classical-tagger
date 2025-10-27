package validation

import (
	"github.com/cehbz/classical-tagger/internal/domain"
)

// RuleMetadata describes a rule's identity and severity
type RuleMetadata struct {
	ID     string
	Name   string
	Level  domain.Level
	Weight float64
}

// RuleResult is what every rule method returns
type RuleResult struct {
	Meta   RuleMetadata
	Issues []domain.ValidationIssue
}

// Passed returns true if the rule found no issues
func (r RuleResult) Passed() bool {
	return len(r.Issues) == 0
}
