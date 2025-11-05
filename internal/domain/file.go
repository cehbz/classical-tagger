package domain

type FileLike interface {
	GetPath() string
}

// File represents a file in a torrent directory.
// All fields are exported and mutable.
type File struct {
	Path string `json:"path"` // Relative path from torrent root
}

func (f *File) GetPath() string {
	return f.Path
}
