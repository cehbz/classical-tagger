package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// PerformerFormat checks that artist tags follow classical format (classical.artist_name)
// Format: "Soloist(s), Orchestra(s)/Ensemble(s), Conductor"
func (r *Rules) PerformerFormat(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.artist_name",
		Name:   "Performer format: Soloist, Ensemble, Conductor",
		Level:  domain.LevelWarning, // Warning because order is recommended but not strictly required
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	for _, track := range actual.Tracks {
		artists := track.Artists
		if len(artists) == 0 {
			continue // Will be caught by RequiredTags rule
		}

		// Extract artists by role
		var soloists []domain.Artist
		var ensembles []domain.Artist
		var conductors []domain.Artist
		var others []domain.Artist

		for _, artist := range artists {
			switch artist.Role {
			case domain.RoleComposer:
				// Skip composer - should not be in Artist tag
				continue
			case domain.RoleSoloist:
				soloists = append(soloists, artist)
			case domain.RoleEnsemble:
				ensembles = append(ensembles, artist)
			case domain.RoleConductor:
				conductors = append(conductors, artist)
			case domain.RoleArranger:
				// Arranger typically credited in title, not artist tag
				continue
			default:
				others = append(others, artist)
			}
		}

		// Check if there are performers at all
		totalPerformers := len(soloists) + len(ensembles) + len(conductors)
		if totalPerformers == 0 {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: No performers in artist tag (found only composer/arranger)",
					formatTrackNumber(track)),
			})
			continue
		}

		// Build expected format
		expectedFormat := formatArtistsByRole(soloists, ensembles, conductors)

		// This is informational - we just want to ensure performers are present
		// The exact order is recommended but the rule states it's not strictly enforced
		if expectedFormat != "" {
			// INFO Level: suggest proper format
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelInfo,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Recommended artist format is: %s",
					formatTrackNumber(track), expectedFormat),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// formatArtistsByRole formats artists according to classical convention
// Format: "Soloist(s), Orchestra(s)/Ensemble(s), Conductor"
func formatArtistsByRole(soloists, ensembles, conductors []domain.Artist) string {
	var parts []string

	// Soloists first
	if len(soloists) > 0 {
		soloistNames := make([]string, len(soloists))
		for i, s := range soloists {
			soloistNames[i] = s.Name
		}
		parts = append(parts, strings.Join(soloistNames, ", "))
	}

	// Ensembles/Orchestras second
	if len(ensembles) > 0 {
		ensembleNames := make([]string, len(ensembles))
		for i, e := range ensembles {
			ensembleNames[i] = e.Name
		}
		parts = append(parts, strings.Join(ensembleNames, ", "))
	}

	// Conductor last
	if len(conductors) > 0 {
		conductorNames := make([]string, len(conductors))
		for i, c := range conductors {
			conductorNames[i] = c.Name
		}
		parts = append(parts, strings.Join(conductorNames, ", "))
	}

	return strings.Join(parts, ", ")
}
