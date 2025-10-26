package validation

import (
	"fmt"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// archiveExtensions lists forbidden archive file extensions
var archiveExtensions = []string{
	".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz",
	".cab", ".ace", ".arj", ".lzh", ".sit", ".sitx",
	".tar.gz", ".tar.bz2", ".tar.xz", ".tgz", ".tbz2", ".txz",
}

// NoArchiveFiles checks that torrent contains no archive files (rule 2.3.1)
func (r *Rules) NoArchiveFiles(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.1",
		name:   "No archive files (zip, rar, etc.) in torrent",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	var issues []domain.ValidationIssue
	
	// Check each track filename for archive extensions
	for _, track := range actual.Tracks() {
		fileName := track.Name()
		if fileName == "" {
			continue
		}
		
		fileNameLower := strings.ToLower(fileName)
		
		// Check for archive extensions
		for _, ext := range archiveExtensions {
			if strings.HasSuffix(fileNameLower, ext) {
				issues = append(issues, domain.NewIssue(
					domain.LevelError,
					track.Track(),
					meta.id,
					fmt.Sprintf("Track %s: Archive file found '%s' (archives not allowed in torrents)",
						formatTrackNumber(track), fileName),
				))
				break // Only report once per file
			}
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
