package scraping

import (
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

type ArtistInference struct {
	OriginalText   string
	Artist         domain.Artist
	Reason         string
	Confidence     string
	AlternateRoles []domain.Role
}

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

			return ArtistInference{
				OriginalText: text,
				Artist: domain.Artist{
					Name: text,
					Role: domain.RoleEnsemble,
				},
				Reason:     "keyword: '" + origKeyword + "' indicates ensemble",
				Confidence: confidence,
			}
		}
	}

	// Check for titles (medium confidence for conductor)
	for _, title := range titleKeywords {
		if strings.HasPrefix(lowerText, title+" ") {
			return ArtistInference{
				OriginalText: text,
				Artist: domain.Artist{
					Name: text,
					Role: domain.RoleConductor,
				},
				Reason:         "title '" + title + "' suggests conductor or notable performer",
				Confidence:     "medium",
				AlternateRoles: []domain.Role{domain.RoleSoloist},
			}
		}
	}

	// Context: name after ensemble suggests conductor (high confidence)
	if afterEnsemble {
		return ArtistInference{
			OriginalText: text,
			Artist: domain.Artist{
				Name: text,
				Role: domain.RoleConductor,
			},
			Reason:     "positioned after ensemble; typical conductor position",
			Confidence: "high",
		}
	}

	// Default: assume soloist (medium confidence, could be conductor)
	return ArtistInference{
		OriginalText: text,
		Artist: domain.Artist{
			Name: text,
			Role: domain.RoleSoloist,
		},
		Reason:         "default assumption for individual name",
		Confidence:     "medium",
		AlternateRoles: []domain.Role{domain.RoleConductor},
	}
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
		previousWasEnsemble = inference.Artist.Role == domain.RoleEnsemble
	}

	return inferences
}

// InferArtistRoleWithAlternates infers role and provides alternate possibilities.
// Used for ambiguous cases where human verification is recommended.
func InferArtistRoleWithAlternates(text string) ArtistInference {
	inference := InferArtistRole(text)

	// For low/medium confidence, add alternates
	if inference.Confidence != "high" {
		role := inference.Artist.Role

		// Add common alternates based on primary inference
		if role == domain.RoleSoloist {
			inference.AlternateRoles = append(inference.AlternateRoles, domain.RoleConductor)
		} else if role == domain.RoleConductor {
			inference.AlternateRoles = append(inference.AlternateRoles, domain.RoleSoloist)
		}
	}

	return inference
}

// IsLowConfidence returns true if the inference confidence is low or medium.
func IsLowConfidence(inference ArtistInference) bool {
	return inference.Confidence == "low" || inference.Confidence == "medium"
}

// FormatInferenceForJSON formats an ArtistInference for JSON parsing notes.
func FormatInferenceForJSON(inference ArtistInference) map[string]any {
	result := map[string]any{
		"text":       inference.OriginalText,
		"name":       inference.Artist.Name,
		"role":       inference.Artist.Role,
		"confidence": inference.Confidence,
		"reason":     inference.Reason,
	}

	if len(inference.AlternateRoles) > 0 {
		result["alternate_roles"] = inference.AlternateRoles
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
