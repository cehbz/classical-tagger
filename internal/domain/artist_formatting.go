package domain

import "strings"

// FormatArtists formats a list of artists according to classical music conventions.
// Format: "Soloist(s), Orchestra/Ensemble, Conductor"
// Composers are excluded from the ARTIST tag.
func FormatArtists(artists []Artist) string {
	if len(artists) == 0 {
		return ""
	}

	var soloists []string
	var ensembles []string
	var conductors []string

	for _, artist := range artists {
		switch artist.Role {
		case RoleSoloist:
			soloists = append(soloists, artist.Name)
		case RoleEnsemble:
			ensembles = append(ensembles, artist.Name)
		case RoleConductor:
			conductors = append(conductors, artist.Name)
		case RoleComposer:
			// Composers excluded from ARTIST tag
			continue
		}
	}

	// Build in order: soloists, ensembles, conductors
	var parts []string
	parts = append(parts, soloists...)
	parts = append(parts, ensembles...)
	parts = append(parts, conductors...)

	return strings.Join(parts, ", ")
}
