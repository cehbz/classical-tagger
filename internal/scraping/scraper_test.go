package scraping

import (
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
		domain: "example.com",
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
	return contains(url, m.domain)
}

func (m *MockExtractor) Extract(url string) (*AlbumData, error) {
	m.callCount++
	if m.shouldError {
		return nil, ErrExtractionFailed
	}
	
	return &AlbumData{
		Title:        "Mock Album",
		OriginalYear: 2020,
		Tracks: []TrackData{
			{
				Disc:     1,
				Track:    1,
				Title:    "Mock Track",
				Composer: "Mock Composer",
			},
		},
	}, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
	       len(s) > len(substr) && containsAfter(s[1:], substr)
}

func containsAfter(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
