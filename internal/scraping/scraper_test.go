package scraping

import (
	"strings"
	"testing"
)

func TestExtractorRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test registration
	extractor := &MockExtractor{domain: "test.com"}
	registry.Register(extractor)

	// Test retrieval
	found := registry.Get("https://test.com/album/123")
	if found == nil {
		t.Error("Expected to find registered extractor")
	}

	// Test no match
	notFound := registry.Get("https://unknown.com/album")
	if notFound != nil {
		t.Error("Expected nil for unknown domain")
	}
}

func TestMockExtractor(t *testing.T) {
	extractor := &MockExtractor{
		domain:      "example.com",
		shouldError: false,
	}

	if !extractor.CanHandle("https://example.com/album/123") {
		t.Error("Should handle example.com URLs")
	}

	if extractor.CanHandle("https://other.com/album") {
		t.Error("Should not handle other.com URLs")
	}
}

// MockExtractor is a test implementation
type MockExtractor struct {
	domain      string
	shouldError bool
	callCount   int
}

func (m *MockExtractor) Name() string {
	return "Mock Extractor"
}

func (m *MockExtractor) CanHandle(url string) bool {
	return strings.Contains(url, m.domain)
}

// MockExtractor
func (m *MockExtractor) Extract(url string) (*ExtractionResult, error) {
	m.callCount++
	if m.shouldError {
		return nil, ErrExtractionFailed
	}

	data := &AlbumData{
		Title:        "Mock Album",
		OriginalYear: 2020,
		Tracks:       []TrackData{ /* ... */ },
	}

	return NewExtractionResult(data), nil
}

func TestSynthesizeMissingLabel(t *testing.T) {
	data := &AlbumData{
		Edition: &EditionData{CatalogNumber: "HMC902170"},
	}
	
	synthesized := SynthesizeMissingEditionData(data)
	
	if !synthesized {
		t.Error("Expected synthesis to occur")
	}
	if data.Edition.Label != "[Unknown Label]" {
		t.Errorf("Label = %q, want %q", data.Edition.Label, "[Unknown Label]")
	}
}

func TestInferLabelFromCatalog(t *testing.T) {
	tests := []struct {
		catalog string
		want    string
	}{
		{"HMC902170", "harmonia mundi"},
		{"DG 479 1234", "Deutsche Grammophon"},
		{"BIS-2345", "BIS Records"},
		{"XYZ12345", ""},
	}

	for _, tt := range tests {
		got := InferLabelFromCatalog(tt.catalog)
		if got != tt.want {
			t.Errorf("InferLabelFromCatalog(%q) = %q, want %q", tt.catalog, got, tt.want)
		}
	}
}