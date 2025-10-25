package scraping

// ExtractionError represents a single error encountered during metadata extraction.
// It is immutable after creation.
type ExtractionError struct {
	field    string
	message  string
	required bool
}

// NewExtractionError creates a new immutable ExtractionError.
func NewExtractionError(field, message string, required bool) ExtractionError {
	return ExtractionError{
		field:    field,
		message:  message,
		required: required,
	}
}

// Field returns the field name where the error occurred.
func (e ExtractionError) Field() string {
	return e.field
}

// Message returns the error message.
func (e ExtractionError) Message() string {
	return e.message
}

// Required returns true if this is a required field error (fatal).
func (e ExtractionError) Required() bool {
	return e.required
}

// ExtractionResult wraps extracted album data with validation errors and warnings.
// It is immutable - all modification methods return new instances.
type ExtractionResult struct {
	data         *AlbumData
	errors       []ExtractionError
	warnings     []string
	parsingNotes map[string]interface{}
}

// NewExtractionResult creates a new ExtractionResult with the given data and no errors.
func NewExtractionResult(data *AlbumData) *ExtractionResult {
	return &ExtractionResult{
		data:         data,
		errors:       make([]ExtractionError, 0),
		warnings:     make([]string, 0),
		parsingNotes: nil,
	}
}

// Data returns the extracted album data.
func (r *ExtractionResult) Data() *AlbumData {
	return r.data
}

// Errors returns a copy of all extraction errors.
func (r *ExtractionResult) Errors() []ExtractionError {
	// Return a copy to maintain immutability
	result := make([]ExtractionError, len(r.errors))
	copy(result, r.errors)
	return result
}

// Warnings returns a copy of all warnings.
func (r *ExtractionResult) Warnings() []string {
	// Return a copy to maintain immutability
	result := make([]string, len(r.warnings))
	copy(result, r.warnings)
	return result
}

// ParsingNotes returns a deep copy of parsing notes, or nil if none exist.
func (r *ExtractionResult) ParsingNotes() map[string]interface{} {
	if r.parsingNotes == nil {
		return nil
	}
	
	// Deep copy the map
	result := make(map[string]interface{}, len(r.parsingNotes))
	for k, v := range r.parsingNotes {
		result[k] = v
	}
	return result
}

// HasErrors returns true if there are any errors (required or optional).
func (r *ExtractionResult) HasErrors() bool {
	return len(r.errors) > 0
}

// HasRequiredErrors returns true if there are any required field errors.
func (r *ExtractionResult) HasRequiredErrors() bool {
	for _, err := range r.errors {
		if err.Required() {
			return true
		}
	}
	return false
}

// WithError returns a new ExtractionResult with the given error added.
// The original result is not modified.
func (r *ExtractionResult) WithError(err ExtractionError) *ExtractionResult {
	newErrors := make([]ExtractionError, len(r.errors)+1)
	copy(newErrors, r.errors)
	newErrors[len(r.errors)] = err
	
	return &ExtractionResult{
		data:         r.data,
		errors:       newErrors,
		warnings:     r.warnings,
		parsingNotes: r.parsingNotes,
	}
}

// WithWarning returns a new ExtractionResult with the given warning added.
// The original result is not modified.
func (r *ExtractionResult) WithWarning(warning string) *ExtractionResult {
	newWarnings := make([]string, len(r.warnings)+1)
	copy(newWarnings, r.warnings)
	newWarnings[len(r.warnings)] = warning
	
	return &ExtractionResult{
		data:         r.data,
		errors:       r.errors,
		warnings:     newWarnings,
		parsingNotes: r.parsingNotes,
	}
}

// WithParsingNotes returns a new ExtractionResult with the given parsing notes.
// The notes are deep copied to maintain immutability.
// The original result is not modified.
func (r *ExtractionResult) WithParsingNotes(notes map[string]interface{}) *ExtractionResult {
	// Deep copy the notes
	notesCopy := make(map[string]interface{}, len(notes))
	for k, v := range notes {
		notesCopy[k] = v
	}
	
	return &ExtractionResult{
		data:         r.data,
		errors:       r.errors,
		warnings:     r.warnings,
		parsingNotes: notesCopy,
	}
}

// ArtistInference represents the result of inferring an artist's role from text.
// It is immutable after creation.
type ArtistInference struct {
	originalText   string
	parsedName     string
	inferredRole   string
	confidence     string
	reason         string
	alternateRoles []string
}

// NewArtistInference creates a new immutable ArtistInference.
func NewArtistInference(originalText, parsedName, inferredRole, confidence string) ArtistInference {
	return ArtistInference{
		originalText:   originalText,
		parsedName:     parsedName,
		inferredRole:   inferredRole,
		confidence:     confidence,
		reason:         "",
		alternateRoles: make([]string, 0),
	}
}

// OriginalText returns the original text that was parsed.
func (a ArtistInference) OriginalText() string {
	return a.originalText
}

// ParsedName returns the extracted artist name.
func (a ArtistInference) ParsedName() string {
	return a.parsedName
}

// InferredRole returns the inferred role (e.g., "ensemble", "conductor").
func (a ArtistInference) InferredRole() string {
	return a.inferredRole
}

// Confidence returns the confidence level ("high", "medium", "low").
func (a ArtistInference) Confidence() string {
	return a.confidence
}

// Reason returns the explanation for why this role was inferred.
func (a ArtistInference) Reason() string {
	return a.reason
}

// AlternateRoles returns a copy of possible alternate roles.
func (a ArtistInference) AlternateRoles() []string {
	result := make([]string, len(a.alternateRoles))
	copy(result, a.alternateRoles)
	return result
}

// WithReason returns a new ArtistInference with the given reason set.
func (a ArtistInference) WithReason(reason string) ArtistInference {
	return ArtistInference{
		originalText:   a.originalText,
		parsedName:     a.parsedName,
		inferredRole:   a.inferredRole,
		confidence:     a.confidence,
		reason:         reason,
		alternateRoles: a.alternateRoles,
	}
}

// WithAlternateRole returns a new ArtistInference with an additional alternate role.
func (a ArtistInference) WithAlternateRole(role string) ArtistInference {
	newRoles := make([]string, len(a.alternateRoles)+1)
	copy(newRoles, a.alternateRoles)
	newRoles[len(a.alternateRoles)] = role
	
	return ArtistInference{
		originalText:   a.originalText,
		parsedName:     a.parsedName,
		inferredRole:   a.inferredRole,
		confidence:     a.confidence,
		reason:         a.reason,
		alternateRoles: newRoles,
	}
}
