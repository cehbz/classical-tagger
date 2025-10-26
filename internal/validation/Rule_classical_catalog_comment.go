package validation

import (
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// CatalogInfoInComments checks for catalog information (classical.catalog_comment)
// INFO level - suggests including label and catalog number in comment field
func (r *Rules) CatalogInfoInComments(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "classical.catalog_comment",
		name:   "Catalog information recommended in comment field",
		level:  domain.LevelInfo,
		weight: 0.1,
	}
	
	var issues []domain.ValidationIssue
	
	// Check if album has edition information
	edition := actual.Edition()
	
	if edition == nil {
		// No edition info - suggest adding it
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			"Consider adding record label and catalog number information",
		))
		return meta.Fail(issues...)
	}
	
	// Check completeness of edition information
	hasLabel := edition.Label() != ""
	hasCatalog := edition.CatalogNumber() != ""
	hasYear := edition.Year() != 0
	
	var missing []string
	if !hasLabel {
		missing = append(missing, "label")
	}
	if !hasCatalog {
		missing = append(missing, "catalog number")
	}
	if !hasYear {
		missing = append(missing, "release year")
	}
	
	if len(missing) > 0 {
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			fmt.Sprintf("Edition information incomplete - consider adding: %v", missing),
		))
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
