package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// PerformerFormat checks that artist tags follow classical format (classical.artist_name)
// Format: "Soloist(s), Orchestra(s)/Ensemble(s), Conductor"
func (r *Rules) PerformerFormat(actualTrack, refTrack *domain.Track, actualAlbum, refAlbum *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.artist_name",
		Name:   "Performer format: Soloist, Ensemble, Conductor",
		Level:  domain.LevelWarning, // Warning because order is recommended but not strictly required
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	artists := actualTrack.Artists
	if len(artists) == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Extract artists by role
	var soloists []domain.Artist
	var ensembles []domain.Artist
	var conductors []domain.Artist
	var others []domain.Artist

	for _, artist := range artists {
		switch artist.Role {
		case domain.RoleComposer:
			// Composer presence is allowed; do not count as a performer
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
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: No performers in artist tag",
				formatTrackNumber(actualTrack)),
		})
		return RuleResult{Meta: meta, Issues: issues}
	}

	// Do not enforce title to contain artist string; only ensure performers exist
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
