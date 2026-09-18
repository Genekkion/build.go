package vfs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// OSFS implements FS using the host operating system.
type OSFS struct{}

// NewOSFS creates an OS-backed file system.
func NewOSFS() *OSFS {
	return &OSFS{}
}

// Open opens a file on the host filesystem.
func (o *OSFS) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Glob matches files on the host filesystem using filepath.Glob.
func (o *OSFS) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Stat returns FileInfo for a file on the host filesystem.
func (o *OSFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}
