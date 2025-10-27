package scraping

import (
	"strings"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestExtractorRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test registration
	extractor := &MockExtractor{Domain: "test.com"}
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
		Domain:      "example.com",
		ShouldError: false,
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
	Domain      string
	ShouldError bool
	CallCount   int
}

func (m *MockExtractor) Name() string {
	return "Mock Extractor"
}

func (m *MockExtractor) CanHandle(url string) bool {
	return strings.Contains(url, m.Domain)
}

// MockExtractor
func (m *MockExtractor) Extract(url string) (*ExtractionResult, error) {
	m.CallCount++
	if m.ShouldError {
		return nil, ErrExtractionFailed
	}

	data := &domain.Album{
		Title:        "Mock Album",
		OriginalYear: 2020,
		Tracks:       []*domain.Track{ /* ... */ },
	}

	return &ExtractionResult{
		Album: data,
	}, nil
}

func TestSynthesizeMissingLabel(t *testing.T) {
	data := &domain.Album{
		Edition: &domain.Edition{
			CatalogNumber: "HMC902170",
		},
	}

	synthesized := SynthesizeMissingEditionData(data)

	if !synthesized {
		t.Error("Expected synthesis to occur")
	}
	if data.Edition.Label != "[Unknown Label]" {
		t.Errorf("Label = %q, want %q", data.Edition.Label, "[Unknown Label]")
	}
}
