package domain

type FileLike interface {
	GetPath() string
	GetSize() int64
}

// File represents a file in a torrent directory.
// All fields are exported and mutable.
type File struct {
	Path string `json:"path"` // Relative path from torrent root
	Size int64  `json:"size"` // File size in bytes
}

func (f *File) GetPath() string {
	return f.Path
}

func (f *File) GetSize() int64 {
	return f.Size
}