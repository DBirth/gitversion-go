package fs

import (
	"io/fs"
	"os"
)

// Filesystem is an interface that abstracts filesystem operations.
type Filesystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Exists(name string) (bool, error)
}

// OsFs is a filesystem implementation that uses the os package.
type OsFs struct{}

// NewOsFs creates a new OsFs.
func NewOsFs() *OsFs {
	return &OsFs{}
}

// ReadFile reads the file named by filename and returns the contents.
func (fs *OsFs) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile writes data to a file named by filename.
func (fs *OsFs) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Exists checks if a file or directory exists.
func (fs *OsFs) Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
