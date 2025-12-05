package domain

import (
	"fmt"
	"strings"
)

// Role represents the role of an artist in a performance.
type Role int

// in display order
const (
	RoleUnknown Role = iota
	RoleComposer
	RoleConductor
	RoleEnsemble
	RoleSoloist
	RolePerformer
	RoleGuest
	RoleDJ
	RoleProducer
	RoleArranger
	RoleRemixer
)

func (r Role) IsPerformer() bool {
	return r == RoleSoloist || r == RoleEnsemble || r == RolePerformer || r == RoleGuest || r == RoleConductor
}

// String returns the lowercase string representation of the role.
func (r Role) String() string {
	switch r {
	case RoleComposer:
		return "composer"
	case RoleConductor:
		return "conductor"
	case RoleEnsemble:
		return "ensemble"
	case RoleSoloist:
		return "soloist"
	case RolePerformer:
		return "performer"
	case RoleGuest:
		return "guest"
	case RoleDJ:
		return "dj"
	case RoleProducer:
		return "producer"
	case RoleArranger:
		return "arranger"
	case RoleRemixer:
		return "remixer"
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
	case "dj":
		return RoleDJ, nil
	case "producer":
		return RoleProducer, nil
	case "remixer":
		return RoleRemixer, nil
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
