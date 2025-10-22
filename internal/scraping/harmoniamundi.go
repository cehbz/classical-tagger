package scraping

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HarmoniaMundiExtractor extracts album data from Harmonia Mundi website.
type HarmoniaMundiExtractor struct {
	client *http.Client
	parser *HarmoniaMundiParser
}

// NewHarmoniaMundiExtractor creates a new Harmonia Mundi extractor.
func NewHarmoniaMundiExtractor() *HarmoniaMundiExtractor {
	return &HarmoniaMundiExtractor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		parser: NewHarmoniaMundiParser(),
	}
}

// Name returns the extractor name.
func (e *HarmoniaMundiExtractor) Name() string {
	return "Harmonia Mundi"
}

// CanHandle returns true if this URL is from Harmonia Mundi.
func (e *HarmoniaMundiExtractor) CanHandle(url string) bool {
	return strings.Contains(url, "harmoniamundi.com")
}

// Extract extracts album data from a Harmonia Mundi URL.
// Returns ExtractionResult with errors and warnings.
func (e *HarmoniaMundiExtractor) Extract(url string) (*ExtractionResult, error) {
	// Fetch the page
	resp, err := e.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	
	// Parse HTML using the parser
	return e.parser.Parse(string(body))
}