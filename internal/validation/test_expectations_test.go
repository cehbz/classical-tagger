package validation

// TrackExpectation represents expected issue counts for a single track
type TrackExpectation struct {
	Errors   int
	Warnings int
	Info     int
}

// CaseExpectation is a slice of per-track expectations (index == track index)
type CaseExpectation []TrackExpectation
