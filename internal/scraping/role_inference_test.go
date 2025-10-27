package scraping

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestInferArtistRole tests role inference from artist text
func TestInferArtistRole(t *testing.T) {
	tests := []struct {
		Name            string
		text            string
		wantRole        domain.Role
		wantConfidence  string
		wantReasonMatch string // substring that should appear in reason
	}{
		{
			Name:            "Kammerchor indicates ensemble",
			text:            "RIAS Kammerchor",
			wantRole:        domain.RoleEnsemble,
			wantConfidence:  "high",
			wantReasonMatch: "Kammerchor",
		},
		{
			Name:            "Orchestra indicates ensemble",
			text:            "Berlin Philharmonic Orchestra",
			wantRole:        domain.RoleEnsemble,
			wantConfidence:  "high",
			wantReasonMatch: "Orchestra",
		},
		{
			Name:            "Quartet indicates ensemble",
			text:            "Emerson String Quartet",
			wantRole:        domain.RoleEnsemble,
			wantConfidence:  "high",
			wantReasonMatch: "Quartet",
		},
		{
			Name:            "Choir indicates ensemble",
			text:            "Monteverdi Choir",
			wantRole:        domain.RoleEnsemble,
			wantConfidence:  "high",
			wantReasonMatch: "Choir",
		},
		{
			Name:            "Ensemble keyword",
			text:            "Academy of Ancient Music",
			wantRole:        domain.RoleEnsemble,
			wantConfidence:  "medium",
			wantReasonMatch: "Academy",
		},
		{
			Name:            "Simple name suggests soloist",
			text:            "Glenn Gould",
			wantRole:        domain.RoleSoloist,
			wantConfidence:  "medium",
			wantReasonMatch: "default",
		},
		{
			Name:            "Title suggests conductor",
			text:            "Sir John Eliot Gardiner",
			wantRole:        domain.RoleConductor,
			wantConfidence:  "medium",
			wantReasonMatch: "title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			inference := InferArtistRole(tt.text)

			if inference.Artist.Role != tt.wantRole {
				t.Errorf("InferredRole = %q, want %q", inference.Artist.Role, tt.wantRole)
			}

			if inference.Confidence != tt.wantConfidence {
				t.Errorf("Confidence = %q, want %q", inference.Confidence, tt.wantConfidence)
			}

			if inference.Artist.Name == "" {
				t.Error("Artist.Name is empty")
			}

			if inference.OriginalText != tt.text {
				t.Errorf("OriginalText = %q, want %q", inference.OriginalText, tt.text)
			}

			if !contains(inference.Reason, tt.wantReasonMatch) {
				t.Errorf("Reason = %q, should contain %q", inference.Reason, tt.wantReasonMatch)
			}
		})
	}
}

// TestInferArtistRoleFromContext tests role inference using contextual information
func TestInferArtistRoleFromContext(t *testing.T) {
	tests := []struct {
		Name           string
		text           string
		afterEnsemble  bool
		wantRole       domain.Role
		wantConfidence string
	}{
		{
			Name:           "name after ensemble suggests conductor",
			text:           "Hans-Christoph Rademann",
			afterEnsemble:  true,
			wantRole:       domain.RoleConductor,
			wantConfidence: "high",
		},
		{
			Name:           "name not after ensemble is less certain",
			text:           "Hans-Christoph Rademann",
			afterEnsemble:  false,
			wantRole:       domain.RoleSoloist,
			wantConfidence: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			inference := InferArtistRoleWithContext(tt.text, tt.afterEnsemble)

			if inference.Artist.Role != tt.wantRole {
				t.Errorf("InferredRole = %q, want %q", inference.Artist.Role, tt.wantRole)
			}

			if inference.Confidence != tt.wantConfidence {
				t.Errorf("Confidence = %q, want %q", inference.Confidence, tt.wantConfidence)
			}
		})
	}
}

// TestParseArtistList tests parsing a comma-separated artist list
func TestParseArtistList(t *testing.T) {
	tests := []struct {
		Name           string
		text           string
		wantCount      int
		wantFirstName  string
		wantFirstRole  domain.Role
		wantSecondRole domain.Role
	}{
		{
			Name:           "ensemble, conductor pattern",
			text:           "RIAS Kammerchor, Hans-Christoph Rademann",
			wantCount:      2,
			wantFirstName:  "RIAS Kammerchor",
			wantFirstRole:  domain.RoleEnsemble,
			wantSecondRole: domain.RoleConductor,
		},
		{
			Name:           "orchestra, conductor pattern",
			text:           "Berlin Philharmonic, Herbert von Karajan",
			wantCount:      2,
			wantFirstName:  "Berlin Philharmonic",
			wantFirstRole:  domain.RoleEnsemble,
			wantSecondRole: domain.RoleConductor,
		},
		{
			Name:          "single artist",
			text:          "Martha Argerich",
			wantCount:     1,
			wantFirstName: "Martha Argerich",
			wantFirstRole: domain.RoleSoloist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			inferences := ParseArtistList(tt.text)

			if len(inferences) != tt.wantCount {
				t.Fatalf("ParseArtistList() returned %d artists, want %d", len(inferences), tt.wantCount)
			}

			if inferences[0].Artist.Name != tt.wantFirstName {
				t.Errorf("First artist name = %q, want %q", inferences[0].Artist.Name, tt.wantFirstName)
			}

			if inferences[0].Artist.Role != tt.wantFirstRole {
				t.Errorf("First artist role = %q, want %q", inferences[0].Artist.Role, tt.wantFirstRole)
			}

			if tt.wantCount > 1 && inferences[1].Artist.Role != tt.wantSecondRole {
				t.Errorf("Second artist role = %q, want %q", inferences[1].Artist.Role, tt.wantSecondRole)
			}
		})
	}
}

// TestLowConfidenceWarnings tests that low confidence inferences generate warnings
func TestLowConfidenceWarnings(t *testing.T) {
	// Ambiguous name that could be conductor or soloist
	inference := InferArtistRole("John Smith")

	if inference.Confidence == "high" {
		t.Error("Expected low or medium confidence for ambiguous name")
	}

	// Should have alternate roles suggested
	if len(inference.AlternateRoles) == 0 {
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
