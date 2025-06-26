package fs

import (
	"io"
	"io/fs"
	"os"

	"github.com/go-git/go-billy/v5"
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

// BillyWrappedFs is a filesystem implementation that wraps a billy.Filesystem.
// It implements the Filesystem interface.
type BillyWrappedFs struct {
	fs billy.Filesystem
}

// NewBillyWrappedFs creates a new BillyWrappedFs.
func NewBillyWrappedFs(fs billy.Filesystem) *BillyWrappedFs {
	return &BillyWrappedFs{fs: fs}
}

// Exists checks if a file or directory exists.
func (b *BillyWrappedFs) Exists(name string) (bool, error) {
	_, err := b.fs.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ReadFile reads the file named by filename and returns the contents.
func (b *BillyWrappedFs) ReadFile(name string) ([]byte, error) {
	file, err := b.fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// WriteFile writes data to a file named by filename.
func (b *BillyWrappedFs) WriteFile(name string, data []byte, perm fs.FileMode) error {
	file, err := b.fs.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}
