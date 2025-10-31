package validation

import (
	"fmt"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// RecordLabelPresent checks that record label and catalog number are present (classical.record_label)
// These should be in the Edition information
func (r *Rules) RecordLabelPresent(actual, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.record_label",
		Name:   "Record label and catalog number present",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	edition := actual.Edition

	// Check if edition information exists at all
	if edition == nil {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0, // Album-level
			Rule:    meta.ID,
			Message: "Edition information missing (should include record label and catalog number)",
		})
		return RuleResult{Meta: meta, Issues: issues}
	}

	// Check for record label and catalog number
	missingLabel := edition.Label == ""
	missingCatalog := edition.CatalogNumber == ""
	if missingLabel && missingCatalog {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: "Record label and catalog number are missing from edition information",
		})
	} else {
		if missingLabel {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: "Record label is missing from edition information",
			})
		}
		if missingCatalog {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: "Catalog number is missing from edition information",
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// RecordLabelAccuracy checks that record label and catalog match reference (if provided)
func (r *Rules) RecordLabelAccuracy(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.record_label.accuracy",
		Name:   "Record label and catalog number accuracy",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	// Only validate if reference has edition information
	if reference == nil || reference.Edition == nil {
		return RuleResult{Meta: meta, Issues: nil} // No reference data to compare against
	}
	refEdition := reference.Edition

	actualEdition := actual.Edition
	if actualEdition == nil {
		// Missing edition entirely - this is caught by RecordLabelPresent
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	// Compare label if reference has it
	if refEdition.Label != "" && actualEdition.Label != refEdition.Label {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Record label mismatch: got '%s', expected '%s'",
				actualEdition.Label, refEdition.Label),
		})
	}

	// Compare catalog number if reference has it
	if refEdition.CatalogNumber != "" && actualEdition.CatalogNumber != refEdition.CatalogNumber {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Catalog number mismatch: got '%s', expected '%s'",
				actualEdition.CatalogNumber, refEdition.CatalogNumber),
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}
