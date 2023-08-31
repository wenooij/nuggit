package v1alpha

import (
	"io/fs"
	"os"
)

type File struct {
	fs.FS    `json:"-"`
	Root     string `json:"root,omitempty"`
	Filename string `json:"filename,omitempty"`
}

func (f *File) ReadFile() ([]byte, error) {
	if f.Root == "" {
		f.Root = "."
	}
	if f.FS == nil {
		f.FS = os.DirFS(f.Root)
	}
	return fs.ReadFile(f.FS, f.Filename)
}
