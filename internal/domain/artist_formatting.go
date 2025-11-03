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
	var unknowns []string

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
		case RoleUnknown:
			unknowns = append(unknowns, artist.Name)
		}
	}

	// Build in order: soloists, ensembles, conductors
	var parts []string
	parts = append(parts, soloists...)
	parts = append(parts, ensembles...)
	parts = append(parts, conductors...)
	// Append unknown-role artists preserving original relative order among them
	parts = append(parts, unknowns...)

	return strings.Join(parts, ", ")
}
