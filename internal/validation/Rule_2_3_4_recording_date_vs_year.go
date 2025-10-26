package validation

import (
	"fmt"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// RecordingDateVsYear checks recording date vs year field usage (rule 2.3.4)
// INFO level - suggests using recording date for original recordings
func (r *Rules) RecordingDateVsYear(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.4",
		name:   "Recording date vs release year - use recording year for originals",
		level:  domain.LevelInfo,
		weight: 0.1,
	}
	
	var issues []domain.ValidationIssue
	
	year := actual.OriginalYear()
	edition := actual.Edition()
	
	// If we have both album year and edition year
	if year != 0 && edition != nil && edition.Year() != 0 {
		editionYear := edition.Year()
		
		// If edition year is significantly later than album year
		// this might be a reissue
		yearDiff := editionYear - year
		
		if yearDiff > 10 {
			issues = append(issues, domain.NewIssue(
				domain.LevelInfo,
				0,
				meta.id,
				fmt.Sprintf("Large gap between recording year (%d) and edition year (%d) - verify which is correct for Year field",
					year, editionYear),
			))
		} else if yearDiff < 0 {
			// Edition year before recording year - likely error
			issues = append(issues, domain.NewIssue(
				domain.LevelInfo,
				0,
				meta.id,
				fmt.Sprintf("Edition year (%d) is before recording year (%d) - check if correct",
					editionYear, year),
			))
		} else if yearDiff >= 3 && yearDiff <= 10 {
			// Moderate gap - might be delayed release
			issues = append(issues, domain.NewIssue(
				domain.LevelInfo,
				0,
				meta.id,
				fmt.Sprintf("Recording year (%d) differs from edition year (%d) - year field should use recording date",
					year, editionYear),
			))
		}
	}
	
	// If we have reference, check consistency
	if reference != nil && reference.OriginalYear() != 0 {
		refYear := reference.OriginalYear()
		if year != 0 && year != refYear {
			diff := year - refYear
			if diff < 0 {
				diff = -diff
			}
			
			if diff > 2 {
				issues = append(issues, domain.NewIssue(
					domain.LevelInfo,
					0,
					meta.id,
					fmt.Sprintf("Year %d differs from reference %d - verify recording vs release year",
						year, refYear),
				))
			}
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
