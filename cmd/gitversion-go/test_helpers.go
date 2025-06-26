package main

import (
	"io/fs"
	"testing"

	"github.com/go-git/go-git/v5"
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
func (m *MemFs) Exists(name string) (bool, error) {
	return afero.Exists(m.Fs, name)
}

// ReadFile reads the file named by filename and returns the contents.
func (m *MemFs) ReadFile(name string) ([]byte, error) {
	return afero.ReadFile(m.Fs, name)
}

// WriteFile writes data to a file named by filename.
func (m *MemFs) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return afero.WriteFile(m.Fs, name, data, perm)
}

func withMockGit(t *testing.T, r *git.Repository, action func()) {
	gitPlainOpen = func(path string) (*git.Repository, error) { return r, nil }
	defer func() { gitPlainOpen = git.PlainOpen }()
	action()
}
