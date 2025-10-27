package scraping

import (
	"testing"
)

func TestDecodeHTMLEntities(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "standard entities",
			Input: "Bach &amp; Beethoven",
			Want:  "Bach & Beethoven",
		},
		{
			Name:  "quotes",
			Input: "&quot;Goldberg Variations&quot;",
			Want:  `"Goldberg Variations"`,
		},
		{
			Name:  "apostrophe",
			Input: "Bach&#039;s Suite",
			Want:  "Bach's Suite",
		},
		{
			Name:  "malformed UTF-8 - Noël",
			Input: "NoÃ«l",
			Want:  "Noël",
		},
		{
			Name:  "malformed UTF-8 - umlaut",
			Input: "MÃ¼ller",
			Want:  "Müller",
		},
		{
			Name:  "no entities",
			Input: "Plain Text",
			Want:  "Plain Text",
		},
		{
			Name:  "mixed",
			Input: "NoÃ«l &amp; MÃ¼ller",
			Want:  "Noël & Müller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := decodeHTMLEntities(tt.Input)
			if got != tt.Want {
				t.Errorf("decodeHTMLEntities(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "simple tags",
			Input: "<b>Bold</b> and <i>italic</i>",
			Want:  "Bold and italic",
		},
		{
			Name:  "nested tags",
			Input: "<div><span>Text</span></div>",
			Want:  "Text",
		},
		{
			Name:  "tags with attributes",
			Input: `<a href="link">Link</a>`,
			Want:  "Link",
		},
		{
			Name:  "no tags",
			Input: "Plain text",
			Want:  "Plain text",
		},
		{
			Name:  "empty tags",
			Input: "Text <br/> more text",
			Want:  "Text  more text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := stripHTMLTags(tt.Input)
			if got != tt.Want {
				t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestCleanWhitespace(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "multiple spaces",
			Input: "Too    many   spaces",
			Want:  "Too many spaces",
		},
		{
			Name:  "leading and trailing",
			Input: "  trim me  ",
			Want:  "trim me",
		},
		{
			Name:  "tabs and newlines",
			Input: "text\t\twith\ntabs\nand\nnewlines",
			Want:  "text with tabs and newlines",
		},
		{
			Name:  "already clean",
			Input: "clean text",
			Want:  "clean text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := cleanWhitespace(tt.Input)
			if got != tt.Want {
				t.Errorf("cleanWhitespace(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "all caps",
			Input: "GOLDBERG VARIATIONS",
			Want:  "Goldberg Variations",
		},
		{
			Name:  "with articles",
			Input: "THE ART OF FUGUE",
			Want:  "The Art of Fugue",
		},
		{
			Name:  "with prepositions",
			Input: "CONCERTO IN D MAJOR",
			Want:  "Concerto in D Major",
		},
		{
			Name:  "already title case",
			Input: "Symphony No. 5",
			Want:  "Symphony No. 5",
		},
		{
			Name:  "with de/la/von",
			Input: "MUSIC OF LA RUE",
			Want:  "Music of la Rue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := toTitleCase(tt.Input)
			if got != tt.Want {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "tabs to spaces",
			Input: "text\twith\ttabs",
			Want:  "text with tabs",
		},
		{
			Name:  "newlines to spaces",
			Input: "line1\nline2\nline3",
			Want:  "line1 line2 line3",
		},
		{
			Name:  "non-breaking spaces",
			Input: "text\u00A0with\u00A0nbsp",
			Want:  "text with nbsp",
		},
		{
			Name:  "mixed whitespace",
			Input: "text\t\n\r  with   mixed",
			Want:  "text with mixed",
		},
		{
			Name:  "multiple spaces",
			Input: "Too    many   spaces",
			Want:  "Too many spaces",
		},
		{
			Name:  "leading and trailing",
			Input: "  trim me  ",
			Want:  "trim me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := normalizeWhitespace(tt.Input)
			if got != tt.Want {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "complete sanitization",
			Input: "<b>NoÃ«l</b> &amp; <i>MÃ¼ller</i>",
			Want:  "Noël & Müller",
		},
		{
			Name:  "html with entities and whitespace",
			Input: "  <div>Text   with   &nbsp;  spaces</div>  ",
			Want:  "Text with spaces",
		},
		{
			Name:  "complex example",
			Input: "<h1>Bach&#039;s   Goldberg\n\nVariations</h1>",
			Want:  "Bach's Goldberg Variations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := sanitizeText(tt.Input)
			if got != tt.Want {
				t.Errorf("sanitizeText(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestCleanHTMLEntities_LegacyAlias(t *testing.T) {
	// Test that legacy function still works
	input := "NoÃ«l &amp; Christmas"
	want := "Noël & Christmas"
	got := cleanHTMLEntities(input)

	if got != want {
		t.Errorf("cleanHTMLEntities(%q) = %q, want %q", input, got, want)
	}
}
