package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// archiveExtensions lists forbidden archive file extensions
var archiveExtensions = []string{
	".7z", ".ace", ".arj", ".bz2", ".cab", ".gz", ".lzh", ".rar", ".sit", ".sitx", ".tar", ".tbz2", ".tgz", ".txz", ".xz", ".zip",
}

// NoArchiveFiles checks that torrent contains no archive files (rule 2.3.1)
func (r *Rules) NoArchiveFiles(actualTrack, refTrack *domain.Track, actualTorrent, refTorrent *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.1",
		Name:   "No archive files (zip, rar, etc.) in torrent",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	if actualTorrent == nil {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	// Check all files in the torrent for archive extensions
	for _, file := range actualTorrent.Files {
		filePath := file.GetPath()
		if filePath == "" {
			continue
		}

		filePathLower := strings.ToLower(filePath)

		// Check for archive extensions
		for _, ext := range archiveExtensions {
			if strings.HasSuffix(filePathLower, ext) {
				// Determine track number for the issue message
				trackNum := -1
				if actualTrack != nil {
					trackNum = actualTrack.Track
				} else if track, ok := file.(*domain.Track); ok {
					trackNum = track.Track
				}

				issues = append(issues, domain.ValidationIssue{
					Level:   domain.LevelError,
					Track:   trackNum,
					Rule:    meta.ID,
					Message: fmt.Sprintf("Archive file found '%s' (archives not allowed in torrents)", filePath),
				})
				break // Only report once per file
			}
		}
	}

	return RuleResult{Meta: meta, Issues: issues}
}
