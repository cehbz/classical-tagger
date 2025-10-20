package domain

import (
	"fmt"
	"strings"
)

// Role represents the role of an artist in a performance.
type Role int

const (
	RoleComposer Role = iota
	RoleSoloist
	RoleEnsemble
	RoleConductor
	RoleArranger
	RoleGuest
)

// String returns the lowercase string representation of the role.
func (r Role) String() string {
	switch r {
	case RoleComposer:
		return "composer"
	case RoleSoloist:
		return "soloist"
	case RoleEnsemble:
		return "ensemble"
	case RoleConductor:
		return "conductor"
	case RoleArranger:
		return "arranger"
	case RoleGuest:
		return "guest"
	default:
		return "unknown"
	}
}

// ParseRole parses a string into a Role. Case-insensitive.
func ParseRole(s string) (Role, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "composer":
		return RoleComposer, nil
	case "soloist":
		return RoleSoloist, nil
	case "ensemble":
		return RoleEnsemble, nil
	case "conductor":
		return RoleConductor, nil
	case "arranger":
		return RoleArranger, nil
	case "guest":
		return RoleGuest, nil
	default:
		return Role(0), fmt.Errorf("invalid role: %q", s)
	}
}
