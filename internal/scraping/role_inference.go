package scraping

import (
	"strings"
)

// Ensemble keywords that indicate an ensemble/orchestra/choir
var ensembleKeywords = []string{
	"kammerchor", "choir", "choeur", "chor",
	"orchestra", "orchestre", "orchester",
	"quartet", "quartett", "quartetto",
	"ensemble", "philharmonic", "symphony",
	"consort", "collegium", "academy",
	"chamber", "soloists",
}

// Title keywords that suggest conductor or notable soloist
var titleKeywords = []string{
	"sir", "dame", "maestro", "professor",
}

// InferArtistRole infers an artist's role from their name/text.
// Returns an ArtistInference with confidence level and reasoning.
func InferArtistRole(text string) ArtistInference {
	return InferArtistRoleWithContext(text, false)
}

// InferArtistRoleWithContext infers an artist's role with contextual information.
// afterEnsemble indicates if this artist appears after an ensemble in a list.
func InferArtistRoleWithContext(text string, afterEnsemble bool) ArtistInference {
	text = strings.TrimSpace(text)
	lowerText := strings.ToLower(text)

	// Check for ensemble keywords
	for _, keyword := range ensembleKeywords {
		if strings.Contains(lowerText, keyword) {
			// Find the original case keyword in text
			origKeyword := extractOriginalKeyword(text, keyword)

			// Determine confidence based on keyword strength
			confidence := "high"
			if keyword == "academy" || keyword == "chamber" || keyword == "soloists" {
				confidence = "medium"
			}

			return NewArtistInference(text, text, "ensemble", confidence).
				WithReason("keyword: '" + origKeyword + "' indicates ensemble")
		}
	}

	// Check for titles (medium confidence for conductor)
	for _, title := range titleKeywords {
		if strings.HasPrefix(lowerText, title+" ") {
			inference := NewArtistInference(text, text, "conductor", "medium").
				WithReason("title '" + title + "' suggests conductor or notable performer")
			return inference.WithAlternateRole("soloist")
		}
	}

	// Context: name after ensemble suggests conductor (high confidence)
	if afterEnsemble {
		return NewArtistInference(text, text, "conductor", "high").
			WithReason("positioned after ensemble; typical conductor position")
	}

	// Default: assume soloist (medium confidence, could be conductor)
	inference := NewArtistInference(text, text, "soloist", "medium").
		WithReason("default assumption for individual name")
	return inference.WithAlternateRole("conductor")
}

// ParseArtistList parses a comma-separated list of artists and infers roles.
// Artists appearing after ensembles are inferred to be conductors.
func ParseArtistList(text string) []ArtistInference {
	// Split by comma
	parts := strings.Split(text, ",")
	if len(parts) == 0 {
		return nil
	}

	inferences := make([]ArtistInference, 0, len(parts))
	previousWasEnsemble := false

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		inference := InferArtistRoleWithContext(part, previousWasEnsemble)
		inferences = append(inferences, inference)

		// Track if this was an ensemble for next iteration
		previousWasEnsemble = inference.InferredRole() == "ensemble"
	}

	return inferences
}

// InferArtistRoleWithAlternates infers role and provides alternate possibilities.
// Used for ambiguous cases where human verification is recommended.
func InferArtistRoleWithAlternates(text string) ArtistInference {
	inference := InferArtistRole(text)

	// For low/medium confidence, add alternates
	if inference.Confidence() != "high" {
		role := inference.InferredRole()

		// Add common alternates based on primary inference
		if role == "soloist" {
			inference = inference.WithAlternateRole("conductor")
		} else if role == "conductor" {
			inference = inference.WithAlternateRole("soloist")
		}
	}

	return inference
}

// IsLowConfidence returns true if the inference confidence is low or medium.
func IsLowConfidence(inference ArtistInference) bool {
	return inference.Confidence() == "low" || inference.Confidence() == "medium"
}

// FormatInferenceForJSON formats an ArtistInference for JSON parsing notes.
func FormatInferenceForJSON(inference ArtistInference) map[string]interface{} {
	result := map[string]interface{}{
		"text":       inference.OriginalText(),
		"name":       inference.ParsedName(),
		"role":       inference.InferredRole(),
		"confidence": inference.Confidence(),
		"reason":     inference.Reason(),
	}

	if len(inference.AlternateRoles()) > 0 {
		result["alternate_roles"] = inference.AlternateRoles()
	}

	return result
}

// extractOriginalKeyword finds the original-case version of a keyword in text.
func extractOriginalKeyword(text, keyword string) string {
	lowerText := strings.ToLower(text)
	keyword = strings.ToLower(keyword)

	// Find keyword position in lowercase text
	index := strings.Index(lowerText, keyword)
	if index == -1 {
		return keyword
	}

	// Extract from original text with original case
	return text[index : index+len(keyword)]
}
