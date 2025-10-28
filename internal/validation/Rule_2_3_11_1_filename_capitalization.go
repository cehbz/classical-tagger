package validation

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// FilenameCapitalization checks that filenames use proper Title Case (rule 2.3.11.1)
func (r *Rules) FilenameCapitalization(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.11.1",
		Name:   "Filename capitalization must be Title Case",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	for _, track := range actual.Tracks {
		fileName := track.Name
		if fileName == "" {
			continue
		}

		// Extract just the filename (not path)
		parts := strings.Split(fileName, "/")
		justFileName := parts[len(parts)-1]

		// Extract the title portion from filename (after track number)
		matches := filenameTrackPattern.FindStringSubmatch(justFileName)
		if len(matches) < 2 {
			continue // Can't parse filename structure
		}

		fileTitle := matches[1]

		// Check capitalization
		capIssue := checkCapitalization(fileTitle)
		if capIssue != "" {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: %s in fileName: '%s'",
					formatTrackNumber(track), capIssue, justFileName),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// checkCapitalization returns an error message if capitalization is wrong, empty string if OK
func checkCapitalization(title string) string {

	// Accept if it matches strict Title Case or Casual Title Case
	if validTitleCase(title) || validCasualTitleCase(title) {
		return ""
	}
	return "Not Title Case or Casual Title Case"
}

// isAllUppercase checks that there is at least one letter and if all letters are uppercase
func isAllUppercase(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter
}

// isAllLowercase checks that there is at least one letter and if all letters are lowercase
func isAllLowercase(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

// validCasualTitleCase checks if string follows Casual Title Case rules (first letter of each word capitalized)
func validCasualTitleCase(s string) bool {
	for _, tok := range strings.Fields(s) {
		if tok == "" {
			continue
		}
		// Allow acronyms/roman/catalog anywhere
		if isAcronym(tok) || isRomanNumeral(tok) || isCatalogToken(tok) {
			continue
		}
		// First letter uppercase if it's a letter
		r, _ := utf8.DecodeRuneInString(tok)
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
		// Disallow fully uppercase multi-letter tokens (not matched as acronym above)
		if len(tok) >= 2 && isAllUppercase(tok) {
			return false
		}
	}
	return true
}

// validTitleCase returns true if each significant token is capitalized, allowing
// acronyms/initialisms, roman numerals, and catalog tokens.
func isTitleCase(title string) bool { return validTitleCase(title) }

// validTitleCase returns true if the string follows Title Case:
// - First and last token of each segment capitalized
// - Small words lowercase unless segment boundary (first/last)
// - Major words capitalized
// - Acronyms/initialisms, roman numerals, catalog tokens allowed anywhere
func validTitleCase(title string) bool {
	segments := splitOnDelimiters(title)
	for _, seg := range segments {
		tokens := strings.Fields(seg)
		if len(tokens) == 0 {
			continue
		}
		for i, tok := range tokens {
			isBoundary := i == 0 || i == len(tokens)-1
			for _, part := range strings.Split(tok, "-") {
				if part == "" {
					continue
				}
				lower := strings.ToLower(part)
				// Accept pure numbers (track/order markers)
				if isNumber(part) {
					continue
				}
				// Accept single-letter key tokens [A-G] and with optional accidental (#/b)
				if isKeyToken(part) {
					continue
				}
				if isAcronym(part) || isRomanNumeral(part) || isCatalogToken(part) {
					continue
				}
				if isSmallWord(lower) && !isBoundary {
					// Must be lowercase for strict Title Case (no exceptions)
					if !isLowercaseWord(part) {
						return false
					}
					continue
				}

				// Accept common classical phrase patterns:
				// - After certain prepositions/articles (con, per, da, di, del, de, der, von, van, y),
				//   allow following word to be lowercase (e.g., "con brio", "per pianoforte").
				if i > 0 {
					prev := strings.ToLower(tokens[i-1])
					switch prev {
					case "con", "per", "da", "di", "del", "de", "der", "von", "van", "y":
						if isLowercaseWord(part) {
							continue
						}
					}
				}

				// - Mode tokens in key phrases (major/minor) may be lowercase: "in D major".
				if (lower == "major" || lower == "minor") && i > 0 {
					// prior token should be the key like C, D, etc., but we keep it simple here
					if isLowercaseWord(part) {
						continue
					}
				}
				if !isCapitalizedWord(part) {
					return false
				}
			}
		}
	}
	return true
}

func splitOnDelimiters(s string) []string {
	// Split on colon and em/en dashes; keep simple to avoid regex overhead
	// Replace delimiters with a unified separator and then split.
	repl := strings.NewReplacer(":", "|", "—", "|", "–", "|", "-", "|")
	unified := repl.Replace(s)
	parts := strings.Split(unified, "|")
	return parts
}

func isSmallWord(w string) bool {
	switch w {
	case "a", "an", "the", "and", "but", "or", "nor", "as", "at", "by", "for", "so", "yet",
		"in", "of", "on", "per", "to", "up", "via", "vs", "vs.",
		"von", "van", "und", "de", "di", "da", "del", "der",
		"la", "le", "les", "du", "des", "el", "y", "con", "non", "troppo":
		return true
	default:
		return false
	}
}

func isAcronym(s string) bool {
	upperSet := map[string]struct{}{
		"LSO": {}, "BBC": {}, "CD": {}, "SACD": {}, "LP": {}, "EP": {}, "DVD": {}, "BD": {}, "UHD": {}, "WEB": {}, "USA": {},
	}
	if _, ok := upperSet[s]; ok {
		return true
	}
	// R&B style
	if strings.ToUpper(s) == s && strings.Contains(s, "&") {
		return true
	}
	// Dotted uppercase like U.S.A.
	dotted := strings.ReplaceAll(s, ".", "")
	if dotted != s && strings.ToUpper(dotted) == dotted && len(dotted) > 1 {
		return true
	}
	return false
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isKeyToken(s string) bool {
	if len(s) == 1 {
		return s[0] >= 'A' && s[0] <= 'G'
	}
	if len(s) == 2 {
		return (s[0] >= 'A' && s[0] <= 'G') && (s[1] == '#' || s[1] == 'b')
	}
	return false
}

func isRomanNumeral(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch r {
		case 'I', 'V', 'X', 'L', 'C', 'D', 'M':
			// ok
		default:
			return false
		}
	}
	return true
}

var catalogTokens = map[string]struct{}{
	"Op.": {}, "No.": {}, "Hob.": {}, "Wq.": {},
	"K.": {}, "D.": {}, "S.": {}, "L.": {}, "P.": {},
	"BWV": {}, "KV": {}, "RV": {}, "HWV": {}, "TWV": {},
}

func isCatalogToken(s string) bool {
	_, ok := catalogTokens[s]
	return ok
}

func isLowercaseWord(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func isCapitalizedWord(s string) bool {
	// Find first letter rune; ensure uppercase. Allow any punctuation/numbers otherwise.
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if unicode.IsLetter(r) {
			if !unicode.IsUpper(r) {
				return false
			}
			// Reject fully uppercase tokens with multiple letters (handled by isAcronym/catelog earlier)
			letters := 1
			j := i + size
			allUpper := true
			for j < len(s) {
				r2, sz2 := utf8.DecodeRuneInString(s[j:])
				if unicode.IsLetter(r2) {
					letters++
					if unicode.IsLower(r2) {
						allUpper = false
					}
				}
				j += sz2
			}
			if letters > 1 && allUpper {
				return false
			}
			return true
		}
		i += size
	}
	// No letters → accept as capitalized (e.g., numbers, punctuation-only tokens)
	return true
}
