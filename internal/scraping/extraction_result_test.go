package scraping

import (
	"testing"
)

// TestExtractionError tests the immutable ExtractionError type
func TestExtractionError(t *testing.T) {
	t.Run("creates error with all fields", func(t *testing.T) {
		err := NewExtractionError("title", "missing required field", true)

		if err.Field() != "title" {
			t.Errorf("Field() = %q, want %q", err.Field(), "title")
		}

		if err.Message() != "missing required field" {
			t.Errorf("Message() = %q, want %q", err.Message(), "missing required field")
		}

		if !err.Required() {
			t.Error("Required() = false, want true")
		}
	})

	t.Run("optional error", func(t *testing.T) {
		err := NewExtractionError("label", "could not parse", false)

		if err.Required() {
			t.Error("Required() = true, want false")
		}
	})
}

// TestExtractionResult tests the immutable ExtractionResult type
func TestExtractionResult(t *testing.T) {
	t.Run("creates result with no errors", func(t *testing.T) {
		data := &AlbumData{
			Title:        "Test Album",
			OriginalYear: 2020,
		}

		result := NewExtractionResult(data)

		if result.Data() != data {
			t.Error("Data() returned different pointer")
		}

		if len(result.Errors()) != 0 {
			t.Errorf("Errors() = %d items, want 0", len(result.Errors()))
		}

		if len(result.Warnings()) != 0 {
			t.Errorf("Warnings() = %d items, want 0", len(result.Warnings()))
		}

		if result.HasErrors() {
			t.Error("HasErrors() = true, want false")
		}

		if result.HasRequiredErrors() {
			t.Error("HasRequiredErrors() = true, want false")
		}
	})

	t.Run("adds errors immutably", func(t *testing.T) {
		data := &AlbumData{Title: "Test"}
		result := NewExtractionResult(data)

		// Add error - should return new instance
		err := NewExtractionError("year", "missing year", true)
		result2 := result.WithError(err)

		// Original should be unchanged
		if len(result.Errors()) != 0 {
			t.Error("Original result was mutated")
		}

		// New result should have error
		if len(result2.Errors()) != 1 {
			t.Errorf("WithError() returned %d errors, want 1", len(result2.Errors()))
		}

		if !result2.HasErrors() {
			t.Error("HasErrors() = false, want true")
		}

		if !result2.HasRequiredErrors() {
			t.Error("HasRequiredErrors() = false, want true")
		}
	})

	t.Run("adds warnings immutably", func(t *testing.T) {
		data := &AlbumData{Title: "Test"}
		result := NewExtractionResult(data)

		result2 := result.WithWarning("low confidence artist detection")

		// Original unchanged
		if len(result.Warnings()) != 0 {
			t.Error("Original result was mutated")
		}

		// New has warning
		if len(result2.Warnings()) != 1 {
			t.Errorf("WithWarning() returned %d warnings, want 1", len(result2.Warnings()))
		}
	})

	t.Run("chains multiple errors", func(t *testing.T) {
		data := &AlbumData{Title: "Test"}
		result := NewExtractionResult(data).
			WithError(NewExtractionError("year", "missing", true)).
			WithError(NewExtractionError("composer", "missing", true)).
			WithWarning("check track 3")

		if len(result.Errors()) != 2 {
			t.Errorf("Chained errors = %d, want 2", len(result.Errors()))
		}

		if len(result.Warnings()) != 1 {
			t.Errorf("Chained warnings = %d, want 1", len(result.Warnings()))
		}
	})

	t.Run("distinguishes required vs optional errors", func(t *testing.T) {
		data := &AlbumData{Title: "Test"}
		result := NewExtractionResult(data).
			WithError(NewExtractionError("label", "missing", false)).
			WithError(NewExtractionError("year", "missing", true))

		if !result.HasErrors() {
			t.Error("HasErrors() = false, want true")
		}

		if !result.HasRequiredErrors() {
			t.Error("HasRequiredErrors() = false, want true")
		}

		// Test with only optional error
		result2 := NewExtractionResult(data).
			WithError(NewExtractionError("label", "missing", false))

		if !result2.HasErrors() {
			t.Error("HasErrors() = false, want true for optional error")
		}

		if result2.HasRequiredErrors() {
			t.Error("HasRequiredErrors() = true, want false for optional error")
		}
	})

	t.Run("adds parsing notes immutably", func(t *testing.T) {
		data := &AlbumData{Title: "Test"}
		result := NewExtractionResult(data)

		notes := map[string]interface{}{
			"artist_inference": "used pattern matching",
		}

		result2 := result.WithParsingNotes(notes)

		if result.ParsingNotes() != nil {
			t.Error("Original result was mutated")
		}

		if result2.ParsingNotes() == nil {
			t.Error("WithParsingNotes() did not set notes")
		}

		// Verify deep copy (changes to original map don't affect result)
		notes["new_key"] = "new_value"
		if _, exists := result2.ParsingNotes()["new_key"]; exists {
			t.Error("Parsing notes were not deep copied")
		}
	})
}

// TestArtistInference tests the immutable ArtistInference type
func TestArtistInference(t *testing.T) {
	t.Run("creates inference with all fields", func(t *testing.T) {
		inf := NewArtistInference(
			"RIAS Kammerchor",
			"RIAS Kammerchor",
			"ensemble",
			"high",
		).WithReason("keyword: 'Kammerchor'")

		if inf.OriginalText() != "RIAS Kammerchor" {
			t.Errorf("OriginalText() = %q, want %q", inf.OriginalText(), "RIAS Kammerchor")
		}

		if inf.ParsedName() != "RIAS Kammerchor" {
			t.Errorf("ParsedName() = %q, want %q", inf.ParsedName(), "RIAS Kammerchor")
		}

		if inf.InferredRole() != "ensemble" {
			t.Errorf("InferredRole() = %q, want %q", inf.InferredRole(), "ensemble")
		}

		if inf.Confidence() != "high" {
			t.Errorf("Confidence() = %q, want %q", inf.Confidence(), "high")
		}

		if inf.Reason() != "keyword: 'Kammerchor'" {
			t.Errorf("Reason() = %q, want %q", inf.Reason(), "keyword: 'Kammerchor'")
		}
	})

	t.Run("adds alternate roles immutably", func(t *testing.T) {
		inf := NewArtistInference("John Smith", "John Smith", "conductor", "medium")

		inf2 := inf.WithAlternateRole("soloist")

		if len(inf.AlternateRoles()) != 0 {
			t.Error("Original inference was mutated")
		}

		if len(inf2.AlternateRoles()) != 1 {
			t.Errorf("WithAlternateRole() = %d roles, want 1", len(inf2.AlternateRoles()))
		}

		if inf2.AlternateRoles()[0] != "soloist" {
			t.Errorf("AlternateRoles()[0] = %q, want %q", inf2.AlternateRoles()[0], "soloist")
		}
	})
}
