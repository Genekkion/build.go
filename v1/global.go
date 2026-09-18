package buildgo

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	slog2 "log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"

	"github.com/Genekkion/build.go/internal/db"
	"github.com/Genekkion/build.go/internal/log/slog"
	"github.com/Genekkion/build.go/internal/vfs"
)

var (
	Logger = func() *slog2.Logger {
		return slog.NewLogger(
			slog.NewHandler(os.Stdout, &slog2.HandlerOptions{
				Level: slog2.LevelInfo,
			}),
		)
	}()
	CacheDir     string
	CacheDb      *sql.DB
	Hasher       = sha256.New
	CurrentFS    vfs.FS = vfs.NewOSFS()
	memDBCounter atomic.Int64
)

type setupConfig struct {
	inMemory bool
	db       *sql.DB
	cacheDir string
	fs       vfs.FS
}

// SetupOption configures the build system during Setup.
type SetupOption func(*setupConfig)

// WithInMemoryDB configures SQLite to run purely in memory without disk access.
func WithInMemoryDB() SetupOption {
	return func(c *setupConfig) {
		c.inMemory = true
	}
}

// WithDB allows providing an existing database connection.
func WithDB(database *sql.DB) SetupOption {
	return func(c *setupConfig) {
		c.db = database
	}
}

// WithCacheDir overrides the default cache directory.
func WithCacheDir(dir string) SetupOption {
	return func(c *setupConfig) {
		c.cacheDir = dir
	}
}

// WithFS overrides the filesystem abstraction used for reading and globbing files.
func WithFS(fileSystem vfs.FS) SetupOption {
	return func(c *setupConfig) {
		c.fs = fileSystem
	}
}

// Setup sets up the global variables and database.
// Warning: will panic if unable to set up successfully.
func Setup(opts ...SetupOption) {
	cfg := setupConfig{
		fs: vfs.NewOSFS(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	CurrentFS = cfg.fs

	var err error
	if cfg.db != nil {
		CacheDb = cfg.db
	} else if cfg.inMemory {
		dbName := fmt.Sprintf("file:buildgo_mem_%d?mode=memory&cache=shared", memDBCounter.Add(1))
		CacheDb, err = db.New(dbName)
		if err != nil {
			panic(err)
		}
	} else {
		err = setupCache(cfg.cacheDir)
		if err != nil {
			panic(err)
		}
	}

	Logger.Debug("Setup complete",
		"cacheDir", CacheDir,
	)
}

// Cleanup cleans up the global database connection.
func Cleanup() {
	if CacheDb != nil {
		_ = CacheDb.Close()
		CacheDb = nil
	}
	CurrentFS = vfs.NewOSFS()
}

// setupCache sets up the cache directory and database.
func setupCache(customDir string) (err error) {
	fp := customDir
	if fp == "" {
		fp = filepath.Join(".", ".gobuild")
	}
	fpAbs, err := filepath.Abs(fp)
	if err != nil {
		Logger.Warn("Unable to use absolute path for cache directory, using relative path instead",
			"error", err,
		)
	} else {
		fp = fpAbs
	}

	Logger.Debug("Using cache directory", "dir", fp)

	CacheDir = fp

	err = os.MkdirAll(fp, 0o755)
	if err != nil {
		return err
	}

	CacheDb, err = db.New(filepath.Join(fp, "cache.db"))
	if err != nil {
		return err
	}

	return nil
}

// GetHash returns the hash for the given step name and file path.
func GetHash(stepName string, fp string) (h []byte, err error) {
	return db.GetHash(CacheDb, stepName, fp)
}

// SetHash sets the hash for the given step name and file path.
func SetHash(stepName string, fp string, h []byte) (err error) {
	return db.SetHash(CacheDb, stepName, fp, h)
}

// hashFile returns the hash of the file contents
func hashFile(fp string) (h []byte, err error) {
	hs := Hasher()

	f, err := CurrentFS.Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 32KB buffer
	buf := make([]byte, 32*1024)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		_, err = hs.Write(buf[:n])
		if err != nil {
			return nil, err
		}
	}

	return hs.Sum(nil), nil
}

// needsRebuild returns the hash of the file if it has changed since the last build
// or nil if it hasn't changed.
func needsRebuild(stepName string, fp string) (h []byte, err error) {
	h, err = hashFile(fp)
	if err != nil {
		return nil, err
	}

	hStored, err := GetHash(stepName, fp)
	if err != nil {
		return nil, err
	}

	if hStored == nil || !slices.Equal(hStored, h) {
		return h, nil
	}
	return nil, nil
}
