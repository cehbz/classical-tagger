package domain

// Album represents a classical music release.
type Album struct {
	FolderName   string    `json:"folder_name"`
	Title        string    `json:"title"`
	OriginalYear int       `json:"original_year"`
	Edition      *Edition  `json:"edition,omitempty"`
	Tracks       []*Track  `json:"tracks"`
}

// IsMultiDisc returns true if the album contains tracks from multiple discs.
// An album is considered multi-disc if any track has Disc > 1 or if there are multiple distinct disc numbers.
func (a *Album) IsMultiDisc() bool {
	if a == nil || len(a.Tracks) == 0 {
		return false
	}

	maxDisc := 1
	discSet := make(map[int]bool)
	for _, track := range a.Tracks {
		if track.Disc > maxDisc {
			maxDisc = track.Disc
		}
		discSet[track.Disc] = true
	}

	// Multi-disc if max disc > 1 OR if there are multiple distinct disc numbers
	return maxDisc > 1 || len(discSet) > 1
}
