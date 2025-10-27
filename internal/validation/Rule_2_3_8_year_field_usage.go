package validation

import (
	"fmt"

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

	year := actual.OriginalYear

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
		if year > 2026 {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Year %d is in the future (use recording/release year)", year),
			})
		}
	}

	// Check against reference if provided
	if reference != nil && reference.OriginalYear != 0 {
		refYear := reference.OriginalYear

		if year != 0 && year != refYear {
			// Calculate difference
			diff := year - refYear
			if diff < 0 {
				diff = -diff
			}

			// Small differences (1-2 years) might be reissue vs original
			level := domain.LevelInfo
			if diff > 5 {
				level = domain.LevelWarning
			}

			issues = append(issues, domain.ValidationIssue{
				Level: level,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Year %d differs from reference %d (difference: %d years)",
					year, refYear, diff),
			})
		}
	}

	// Check Edition year if present
	if actual.Edition != nil {
		editionYear := actual.Edition.Year
		if editionYear != 0 && year != 0 {
			// Edition year should typically be same or later than recording year
			if editionYear < year {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelWarning,
					Track: 0,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Edition year %d is earlier than album year %d (check if correct)",
						editionYear, year),
				})
			}
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
