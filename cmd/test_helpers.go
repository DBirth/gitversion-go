package cmd

import (
	"io/fs"

	"github.com/spf13/afero"
)

// MemFs is a filesystem implementation that uses afero's in-memory filesystem.
// It implements the fs.Filesystem interface.
type MemFs struct {
	afero.Fs
}

// NewMemFs creates a new MemFs.
func NewMemFs() *MemFs {
	return &MemFs{Fs: afero.NewMemMapFs()}
}

// Exists checks if a file or directory exists.
func (mfs *MemFs) Exists(name string) (bool, error) {
	return afero.Exists(mfs, name)
}

// ReadFile reads the file named by filename and returns the contents.
func (mfs *MemFs) ReadFile(name string) ([]byte, error) {
	return afero.ReadFile(mfs, name)
}

// WriteFile writes data to a file named by filename.
func (mfs *MemFs) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return afero.WriteFile(mfs, name, data, perm)
}
