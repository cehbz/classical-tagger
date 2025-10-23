package scraping

import (
	"testing"
)

func TestDecodeHTMLEntities(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "standard entities",
			input: "Bach &amp; Beethoven",
			want:  "Bach & Beethoven",
		},
		{
			name:  "quotes",
			input: "&quot;Goldberg Variations&quot;",
			want:  `"Goldberg Variations"`,
		},
		{
			name:  "apostrophe",
			input: "Bach&#039;s Suite",
			want:  "Bach's Suite",
		},
		{
			name:  "malformed UTF-8 - Noël",
			input: "NoÃ«l",
			want:  "Noël",
		},
		{
			name:  "malformed UTF-8 - umlaut",
			input: "MÃ¼ller",
			want:  "Müller",
		},
		{
			name:  "no entities",
			input: "Plain Text",
			want:  "Plain Text",
		},
		{
			name:  "mixed",
			input: "NoÃ«l &amp; MÃ¼ller",
			want:  "Noël & Müller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeHTMLEntities(tt.input)
			if got != tt.want {
				t.Errorf("decodeHTMLEntities(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple tags",
			input: "<b>Bold</b> and <i>italic</i>",
			want:  "Bold and italic",
		},
		{
			name:  "nested tags",
			input: "<div><span>Text</span></div>",
			want:  "Text",
		},
		{
			name:  "tags with attributes",
			input: `<a href="link">Link</a>`,
			want:  "Link",
		},
		{
			name:  "no tags",
			input: "Plain text",
			want:  "Plain text",
		},
		{
			name:  "empty tags",
			input: "Text <br/> more text",
			want:  "Text  more text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTMLTags(tt.input)
			if got != tt.want {
				t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "multiple spaces",
			input: "Too    many   spaces",
			want:  "Too many spaces",
		},
		{
			name:  "leading and trailing",
			input: "  trim me  ",
			want:  "trim me",
		},
		{
			name:  "tabs and newlines",
			input: "text\t\twith\ntabs\nand\nnewlines",
			want:  "text with tabs and newlines",
		},
		{
			name:  "already clean",
			input: "clean text",
			want:  "clean text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("cleanWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "all caps",
			input: "GOLDBERG VARIATIONS",
			want:  "Goldberg Variations",
		},
		{
			name:  "with articles",
			input: "THE ART OF FUGUE",
			want:  "The Art of Fugue",
		},
		{
			name:  "with prepositions",
			input: "CONCERTO IN D MAJOR",
			want:  "Concerto in D Major",
		},
		{
			name:  "already title case",
			input: "Symphony No. 5",
			want:  "Symphony No. 5",
		},
		{
			name:  "with de/la/von",
			input: "MUSIC OF LA RUE",
			want:  "Music of la Rue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toTitleCase(tt.input)
			if got != tt.want {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "tabs to spaces",
			input: "text\twith\ttabs",
			want:  "text with tabs",
		},
		{
			name:  "newlines to spaces",
			input: "line1\nline2\nline3",
			want:  "line1 line2 line3",
		},
		{
			name:  "non-breaking spaces",
			input: "text\u00A0with\u00A0nbsp",
			want:  "text with nbsp",
		},
		{
			name:  "mixed whitespace",
			input: "text\t\n\r  with   mixed",
			want:  "text with mixed",
		},
		{
			name:  "multiple spaces",
			input: "Too    many   spaces",
			want:  "Too many spaces",
		},
		{
			name:  "leading and trailing",
			input: "  trim me  ",
			want:  "trim me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "complete sanitization",
			input: "<b>NoÃ«l</b> &amp; <i>MÃ¼ller</i>",
			want:  "Noël & Müller",
		},
		{
			name:  "html with entities and whitespace",
			input: "  <div>Text   with   &nbsp;  spaces</div>  ",
			want:  "Text with spaces",
		},
		{
			name:  "complex example",
			input: "<h1>Bach&#039;s   Goldberg\n\nVariations</h1>",
			want:  "Bach's Goldberg Variations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeText(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeText(%q) = %q, want %q", tt.input, got, tt.want)
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
