package scraping

import (
	"testing"
)

// TestInferArtistRole tests role inference from artist text
func TestInferArtistRole(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		wantRole        string
		wantConfidence  string
		wantReasonMatch string // substring that should appear in reason
	}{
		{
			name:            "Kammerchor indicates ensemble",
			text:            "RIAS Kammerchor",
			wantRole:        "ensemble",
			wantConfidence:  "high",
			wantReasonMatch: "Kammerchor",
		},
		{
			name:            "Orchestra indicates ensemble",
			text:            "Berlin Philharmonic Orchestra",
			wantRole:        "ensemble",
			wantConfidence:  "high",
			wantReasonMatch: "Orchestra",
		},
		{
			name:            "Quartet indicates ensemble",
			text:            "Emerson String Quartet",
			wantRole:        "ensemble",
			wantConfidence:  "high",
			wantReasonMatch: "Quartet",
		},
		{
			name:            "Choir indicates ensemble",
			text:            "Monteverdi Choir",
			wantRole:        "ensemble",
			wantConfidence:  "high",
			wantReasonMatch: "Choir",
		},
		{
			name:            "Ensemble keyword",
			text:            "Academy of Ancient Music",
			wantRole:        "ensemble",
			wantConfidence:  "medium",
			wantReasonMatch: "Academy",
		},
		{
			name:            "Simple name suggests soloist",
			text:            "Glenn Gould",
			wantRole:        "soloist",
			wantConfidence:  "medium",
			wantReasonMatch: "default",
		},
		{
			name:            "Title suggests conductor",
			text:            "Sir John Eliot Gardiner",
			wantRole:        "conductor",
			wantConfidence:  "medium",
			wantReasonMatch: "title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inference := InferArtistRole(tt.text)

			if inference.InferredRole() != tt.wantRole {
				t.Errorf("InferredRole() = %q, want %q", inference.InferredRole(), tt.wantRole)
			}

			if inference.Confidence() != tt.wantConfidence {
				t.Errorf("Confidence() = %q, want %q", inference.Confidence(), tt.wantConfidence)
			}

			if inference.ParsedName() == "" {
				t.Error("ParsedName() is empty")
			}

			if inference.OriginalText() != tt.text {
				t.Errorf("OriginalText() = %q, want %q", inference.OriginalText(), tt.text)
			}

			if !contains(inference.Reason(), tt.wantReasonMatch) {
				t.Errorf("Reason() = %q, should contain %q", inference.Reason(), tt.wantReasonMatch)
			}
		})
	}
}

// TestInferArtistRoleFromContext tests role inference using contextual information
func TestInferArtistRoleFromContext(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		afterEnsemble  bool
		wantRole       string
		wantConfidence string
	}{
		{
			name:           "name after ensemble suggests conductor",
			text:           "Hans-Christoph Rademann",
			afterEnsemble:  true,
			wantRole:       "conductor",
			wantConfidence: "high",
		},
		{
			name:           "name not after ensemble is less certain",
			text:           "Hans-Christoph Rademann",
			afterEnsemble:  false,
			wantRole:       "soloist",
			wantConfidence: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inference := InferArtistRoleWithContext(tt.text, tt.afterEnsemble)

			if inference.InferredRole() != tt.wantRole {
				t.Errorf("InferredRole() = %q, want %q", inference.InferredRole(), tt.wantRole)
			}

			if inference.Confidence() != tt.wantConfidence {
				t.Errorf("Confidence() = %q, want %q", inference.Confidence(), tt.wantConfidence)
			}
		})
	}
}

// TestParseArtistList tests parsing a comma-separated artist list
func TestParseArtistList(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		wantCount      int
		wantFirstName  string
		wantFirstRole  string
		wantSecondRole string
	}{
		{
			name:           "ensemble, conductor pattern",
			text:           "RIAS Kammerchor, Hans-Christoph Rademann",
			wantCount:      2,
			wantFirstName:  "RIAS Kammerchor",
			wantFirstRole:  "ensemble",
			wantSecondRole: "conductor",
		},
		{
			name:           "orchestra, conductor pattern",
			text:           "Berlin Philharmonic, Herbert von Karajan",
			wantCount:      2,
			wantFirstName:  "Berlin Philharmonic",
			wantFirstRole:  "ensemble",
			wantSecondRole: "conductor",
		},
		{
			name:          "single artist",
			text:          "Martha Argerich",
			wantCount:     1,
			wantFirstName: "Martha Argerich",
			wantFirstRole: "soloist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inferences := ParseArtistList(tt.text)

			if len(inferences) != tt.wantCount {
				t.Fatalf("ParseArtistList() returned %d artists, want %d", len(inferences), tt.wantCount)
			}

			if inferences[0].ParsedName() != tt.wantFirstName {
				t.Errorf("First artist name = %q, want %q", inferences[0].ParsedName(), tt.wantFirstName)
			}

			if inferences[0].InferredRole() != tt.wantFirstRole {
				t.Errorf("First artist role = %q, want %q", inferences[0].InferredRole(), tt.wantFirstRole)
			}

			if tt.wantCount > 1 && inferences[1].InferredRole() != tt.wantSecondRole {
				t.Errorf("Second artist role = %q, want %q", inferences[1].InferredRole(), tt.wantSecondRole)
			}
		})
	}
}

// TestLowConfidenceWarnings tests that low confidence inferences generate warnings
func TestLowConfidenceWarnings(t *testing.T) {
	// Ambiguous name that could be conductor or soloist
	inference := InferArtistRole("John Smith")

	if inference.Confidence() == "high" {
		t.Error("Expected low or medium confidence for ambiguous name")
	}

	// Should have alternate roles suggested
	if len(inference.AlternateRoles()) == 0 {
		t.Error("Expected alternate roles for low confidence inference")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
