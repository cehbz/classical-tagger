package domain

// Level represents the severity of a validation issue.
type Level int

const (
	LevelError Level = iota
	LevelWarning
	LevelInfo
)

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelError:
		return "ERROR"
	case LevelWarning:
		return "WARNING"
	case LevelInfo:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}
