package domain

import (
	"fmt"
	"strings"
)

// Role represents the role of an artist in a performance.
type Role int

const (
	RoleUnknown Role = iota
	RoleComposer
	RoleSoloist
	RoleEnsemble
	RoleConductor
	RoleArranger
	RoleGuest
	RoleProducer
	RolePerformer
)

func (r Role) IsPerformer() bool {
	return r != RoleComposer && r != RoleArranger && r != RoleUnknown
}

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
	case "unknown":
		return RoleUnknown, nil
	default:
		return RoleUnknown, fmt.Errorf("invalid role: %q", s)
	}
}

// MarshalJSON implements json.Marshaler for Role.
func (r Role) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for Role.
func (r *Role) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := strings.Trim(string(data), `"`)
	role, err := ParseRole(s)
	if err != nil {
		return err
	}
	*r = role
	return nil
}
