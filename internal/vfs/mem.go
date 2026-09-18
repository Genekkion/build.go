package vfs

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type memFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (m memFileInfo) Name() string       { return m.name }
func (m memFileInfo) Size() int64        { return m.size }
func (m memFileInfo) Mode() os.FileMode  { return m.mode }
func (m memFileInfo) ModTime() time.Time { return m.modTime }
func (m memFileInfo) IsDir() bool        { return m.mode.IsDir() }
func (m memFileInfo) Sys() any           { return nil }

// MemFS is an in-memory thread-safe virtual file system.
type MemFS struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewMemFS creates an in-memory file system.
func NewMemFS() *MemFS {
	return &MemFS{
		files: make(map[string][]byte),
	}
}

// WriteFile writes data to an in-memory file.
func (m *MemFS) WriteFile(name string, data []byte, _ os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := filepath.Clean(name)
	cpy := make([]byte, len(data))
	copy(cpy, data)
	m.files[cleaned] = cpy
	return nil
}

// Remove deletes an in-memory file.
func (m *MemFS) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := filepath.Clean(name)
	if _, ok := m.files[cleaned]; !ok {
		return os.ErrNotExist
	}
	delete(m.files, cleaned)
	return nil
}

// Open opens an in-memory file for reading.
func (m *MemFS) Open(name string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cleaned := filepath.Clean(name)
	data, ok := m.files[cleaned]
	if !ok {
		return nil, os.ErrNotExist
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Stat returns FileInfo for an in-memory file.
func (m *MemFS) Stat(name string) (fs.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cleaned := filepath.Clean(name)
	data, ok := m.files[cleaned]
	if !ok {
		return nil, os.ErrNotExist
	}

	return memFileInfo{
		name:    filepath.Base(cleaned),
		size:    int64(len(data)),
		mode:    0o644,
		modTime: time.Now(),
	}, nil
}

// Glob returns the names of all files matching pattern.
func (m *MemFS) Glob(pattern string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cleanedPattern := filepath.Clean(pattern)
	var matches []string

	if _, ok := m.files[cleanedPattern]; ok {
		return []string{cleanedPattern}, nil
	}

	for path := range m.files {
		matched, err := filepath.Match(cleanedPattern, path)
		if err != nil {
			return nil, err
		}
		if matched {
			matches = append(matches, path)
		}
	}

	slices.Sort(matches)
	return matches, nil
}
