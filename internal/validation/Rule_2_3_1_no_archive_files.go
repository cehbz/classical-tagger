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
func (r *Rules) NoArchiveFiles(actualTrack, refTrack *domain.Track, actualAlbum, refAlbum *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.1",
		Name:   "No archive files (zip, rar, etc.) in torrent",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	if actualTrack == nil || actualTrack.Name == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}
	
	var issues []domain.ValidationIssue

	// Check track filename for archive extensions
	fileName := actualTrack.Name
	fileNameLower := strings.ToLower(fileName)

	// Check for archive extensions
	for _, ext := range archiveExtensions {
		if strings.HasSuffix(fileNameLower, ext) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Archive file found '%s' (archives not allowed in torrents)",
					formatTrackNumber(actualTrack), fileName),
			})
			break // Only report once per file
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}
