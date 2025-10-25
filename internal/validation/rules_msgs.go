package validation

// ValidationRules contains the text of validation rules from the requirements documents.
// Rules are indexed by their section number or identifier.
var ValidationRules = map[string]string{
	// Name Formatting Rules (highest priority)
	"2.3.2": "Name your directories with meaningful titles, such as 'Artist - Album (Year) - Format'. The minimum acceptable is 'Album' although it is preferable to include more information.",
	"2.3.3": "Avoid creating unnecessary nested folders inside your properly named directory. For single disc albums, all audio files must be included in the main torrent folder. For multi-disc albums, the main torrent folder may include one sub-folder that holds the audio file contents for each disc.",
	"2.3.6": "Torrent album titles must accurately reflect the actual album titles. Use proper capitalization (Title Case) when naming your albums, match published stylization, or reference a reputable source.",
	"2.3.11": "File names must accurately reflect the song titles. Torrents containing files that are named with incorrect song titles can be trumped by properly labeled torrents.",
	"2.3.11.1": "Spelling, characters and capitalization: Artist names and titles that are misspelled in the filenames are grounds for trumping. Improper capitalization in the filenames is grounds for trumping: this includes tracks which are all capitalized. Proper Title Case or casual Title Case is acceptable.",
	"2.3.12": "The maximum character length for files is 180 characters. Path length values must not be so long that they cause incompatibility problems with operating systems and media players.",
	"2.3.13": "Track numbers are required in file names (e.g., '01 - TrackName.mp3').",
	"2.3.14": "When formatted properly, file names will alphabetically sort into the original playing order of the release. Leading zeroes in file names for track numbers are recommended if the number of tracks go over 9.",
	"2.3.14.1": "For albums with more than one artist, if the name of the artist is in the file name, it must come after the track number in order for the tracks to sort into the correct order.",
	"2.3.15": "Multiple-disc torrents cannot have tracks with the same numbers in one directory. You may place all the tracks for disc one in one directory and all the tracks for disc two in another directory.",
	"2.3.16.4": "Required tags: Artist, Album, Title, Track Number. Optional tags: Year (strongly encouraged).",
	"2.3.18.3": "Tags with multiple fields in the same tag (e.g., track number and track title in the track title tags) are subject to trumping.",
	"2.3.20": "Leading spaces are not allowed in any file or folder names.",
	
	// Classical Music Guide Rules (clarifications of general rules)
	"classical.composer": "This tag is designed to contain the name of the composer of each track. Names should be complete, but as long as the composer is uniquely identifiable it is not trumpable.",
	"classical.artist_name": "Artist Name tags must be reserved for the performer(s) and conductor(s) of each track. Format: Soloist(s), Orchestra(s)/Ensemble(s), Conductor.",
	"classical.track_title": "Track Title is about the name of the piece, and that alone. Do NOT repeat the name of the composer inside this tag! Works with multiple movements must have the full work title in every track.",
	"classical.opus": "Inclusion of opus or catalog numbers is preferable but their lack is not grounds for trump as long as the work title is present.",
	"classical.arrangement": "If the work in question is an arrangement, credit the arranger by appending a (arr. by _______) to the track title.",
	"classical.year": "A common misconception about this tag is that it should reflect the date of the recording. This is wrong. Use the date of original release.",
	"classical.album_artist": "When the performer(s) do not remain the same throughout all tracks, this tag is used to credit the one who does appear in all tracks.",
	"classical.folder_name": "Mentioning composer(s) and performer(s) in folder names is crucial. Drop first names and use common abbreviations (J.S. Bach for Johann Sebastian Bach, LSO for London Symphony Orchestra).",
	"classical.record_label": "Make sure you fill out the Record label field (Catalog number too, if you can find it).",
}

// GetRule returns the full text of a validation rule by its identifier.
func GetRule(id string) string {
	if rule, ok := ValidationRules[id]; ok {
		return rule
	}
	return id // fallback to just returning the ID if rule not found
}
