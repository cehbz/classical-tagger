package validation

import (
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// RecordLabelPresent checks that record label and catalog number are present (classical.record_label)
// These should be in the Edition information
func (r *Rules) RecordLabelPresent(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "classical.record_label",
		name:   "Record label and catalog number present",
		level:  domain.LevelWarning,
		weight: 0.5,
	}
	
	var issues []domain.ValidationIssue
	
	edition := actual.Edition()
	
	// Check if edition information exists at all
	if edition == nil {
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0, // Album-level
			meta.id,
			"Edition information missing (should include record label and catalog number)",
		))
		return meta.Fail(issues...)
	}
	
	// Check for record label
	if edition.Label() == "" {
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0,
			meta.id,
			"Record label is missing from edition information",
		))
	}
	
	// Check for catalog number
	if edition.CatalogNumber() == "" {
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0,
			meta.id,
			"Catalog number is missing from edition information",
		))
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// RecordLabelAccuracy checks that record label and catalog match reference (if provided)
func (r *Rules) RecordLabelAccuracy(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "classical.record_label.accuracy",
		name:   "Record label and catalog number accuracy",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	// Only validate if reference has edition information
	refEdition := reference.Edition()
	if refEdition == nil {
		return meta.Pass() // No reference data to compare against
	}
	
	actualEdition := actual.Edition()
	if actualEdition == nil {
		// Missing edition entirely - this is caught by RecordLabelPresent
		return meta.Pass()
	}
	
	var issues []domain.ValidationIssue
	
	// Compare label if reference has it
	if refEdition.Label() != "" && actualEdition.Label() != refEdition.Label() {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			0,
			meta.id,
			fmt.Sprintf("Record label mismatch: got '%s', expected '%s'",
				actualEdition.Label(), refEdition.Label()),
		))
	}
	
	// Compare catalog number if reference has it
	if refEdition.CatalogNumber() != "" && actualEdition.CatalogNumber() != refEdition.CatalogNumber() {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			0,
			meta.id,
			fmt.Sprintf("Catalog number mismatch: got '%s', expected '%s'",
				actualEdition.CatalogNumber(), refEdition.CatalogNumber()),
		))
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
