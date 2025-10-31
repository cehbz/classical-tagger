package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumTitleAccuracy checks that album title matches reference (rule 2.3.6)
// Uses fuzzy matching to allow for minor differences
func (r *Rules) AlbumTitleAccuracy(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.6",
		Name:   "Album title must accurately match reference",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	// Only validate if reference is provided
	if reference == nil {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	actualTitle := actual.Title
	referenceTitle := reference.Title

	if actualTitle == "" || referenceTitle == "" {
		return RuleResult{Meta: meta, Issues: nil} // Will be caught by RequiredTags
	}

	a := clean(actualTitle)
	b := clean(referenceTitle)

	na := normalizeTitle(a)
	nb := normalizeTitle(b)

	if na == nb {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Allow abbreviation to core work name like "Symphony No. 5"
	coreWorkRe := regexp.MustCompile(`(?i)^\s*([a-z]+\s+no\.?\s*\d+)\b`)
	am := coreWorkRe.FindStringSubmatch(na)
	bm := coreWorkRe.FindStringSubmatch(nb)
	if len(am) > 1 && len(bm) > 1 && am[1] == bm[1] {
		// Exact core work matches; if actual is only the core, accept
		if na == am[1] {
			return RuleResult{Meta: meta, Issues: nil}
		}
		// If actual ends with " in <key>" while ref has " in <key> (Major|Minor) ..." → warning
		keyOnly := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(am[1]) + `\s+in\s+[a-g][b#]?$`)
		keyWithMode := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(am[1]) + `\s+in\s+[a-g][b#]?\s+(major|minor)\b`)
		if keyOnly.MatchString(na) && keyWithMode.MatchString(nb) {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Album title '%s' differs from reference '%s' (missing mode/key details)", actualTitle, referenceTitle),
			})
			return RuleResult{Meta: meta, Issues: issues}
		}
	}

	// If both contain explicit musical key with mode, and they differ → error
	keyModeRe := regexp.MustCompile(`(?i)\bin\s+([a-g][b#]?)\s+(major|minor)\b`)
	ak := keyModeRe.FindStringSubmatch(na)
	bk := keyModeRe.FindStringSubmatch(nb)
	if len(ak) >= 3 && len(bk) >= 3 {
		if !strings.EqualFold(ak[1], bk[1]) || !strings.EqualFold(ak[2], bk[2]) {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelError,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Album title '%s' does not match reference '%s' (different key)", actualTitle, referenceTitle),
			})
			return RuleResult{Meta: meta, Issues: issues}
		}
	}

	// Fallback to distance thresholds
	distance := levenshteinDistance(na, nb)
	if distance > 10 { // TODO: extract magic constant
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title '%s' does not match reference '%s'", actualTitle, referenceTitle),
		})
	} else if distance > 3 {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title '%s' differs from reference '%s' (minor differences)", actualTitle, referenceTitle),
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// Strip bracketed extras and composer prefixes for comparison
func clean(s string) string {
	s = regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	// Remove composer prefix patterns like "Beethoven - " or "Beethoven: "
	s = regexp.MustCompile(`^[^\-:]+[\-:]\s*`).ReplaceAllString(s, "")
	return s
}
