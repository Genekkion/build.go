package vfs

import (
	"io"
	"io/fs"
	"os"
)

// FS represents a file system for reading and matching files.
type FS interface {
	Open(name string) (io.ReadCloser, error)
	Glob(pattern string) ([]string, error)
	Stat(name string) (fs.FileInfo, error)
}

// WritableFS extends FS with write operations for testing and mocking.
type WritableFS interface {
	FS
	WriteFile(name string, data []byte, perm os.FileMode) error
	Remove(name string) error
}
