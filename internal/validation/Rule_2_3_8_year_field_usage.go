package validation

import (
	"fmt"
	"time"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// YearFieldUsage checks proper year field usage (rule 2.3.8)
// Recording year for original recordings, release year for reissues
func (r *Rules) YearFieldUsage(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.8",
		Name:   "Year field must use recording year (or original release for reissues)",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	// Compute effective year: prefer OriginalYear; if zero, fall back to Edition.Year if present
	year := actual.OriginalYear
	if year == 0 {
		if actual.Edition != nil && actual.Edition.Year != 0 {
			year = actual.Edition.Year
		} else {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: "Year is missing (should include recording/original year)",
			})
		}
	}

	// Check year is reasonable (not in future, not too old)
	if year != 0 {
		// Classical recordings unlikely before 1900
		if year < 1900 {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Year %d seems too early for a recording (check if correct)", year),
			})
		}

		// Check not in future (allow 1 year ahead for pre-releases)
		if year > time.Now().Year()+1 {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Year %d is in the future (use recording/release year)", year),
			})
		}
	}

	// Check against reference if provided
	if reference != nil {
		refYear := reference.OriginalYear
		if refYear == 0 && reference.Edition != nil {
			refYear = reference.Edition.Year
		}

		if year != 0 && refYear != 0 && year != refYear {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Year %d differs from reference %d",
					year, refYear),
			})
		}
	}

	// Check Edition year if present and we have an explicit album year
	if actual.Edition != nil && actual.OriginalYear != 0 {
		editionYear := actual.Edition.Year
		if editionYear != 0 {
			// Edition year should typically be same or later than recording year
			if editionYear < actual.OriginalYear {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelWarning,
					Track: 0,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Edition year %d is earlier than album year %d (check if correct)",
						editionYear, actual.OriginalYear),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
