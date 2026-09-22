package buildgo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"hash"
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
	memDBCounter atomic.Int64
)

type ctxKeyHasher struct{}

func WithHasher(ctx context.Context, h func() hash.Hash) context.Context {
	return context.WithValue(ctx, ctxKeyHasher{}, h)
}

func HasherFromCtx(ctx context.Context) func() hash.Hash {
	if h, ok := ctx.Value(ctxKeyHasher{}).(func() hash.Hash); ok {
		return h
	}
	return sha256.New
}

type setupConfig struct {
	inMemory bool
	db       *sql.DB
	cacheDir string
	fs       vfs.FS
	logger   *slog2.Logger
	hasher   func() hash.Hash
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

// WithLogger overrides the logger injected into the build context.
func WithLogger(l *slog2.Logger) SetupOption {
	return func(c *setupConfig) {
		c.logger = l
	}
}

// Setup creates a build context with all infrastructure injected.
// Warning: will panic if unable to set up successfully.
func Setup(opts ...SetupOption) context.Context {
	cfg := setupConfig{
		fs:     vfs.NewOSFS(),
		hasher: sha256.New,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx := context.Background()
	ctx = vfs.WithFS(ctx, cfg.fs)

	if cfg.hasher != nil {
		ctx = WithHasher(ctx, cfg.hasher)
	}
	if cfg.logger != nil {
		ctx = slog.WithLogger(ctx, cfg.logger)
	}

	var (
		database *sql.DB
		err      error
	)
	if cfg.db != nil {
		database = cfg.db
	} else if cfg.inMemory {
		dbName := fmt.Sprintf("file:buildgo_mem_%d?mode=memory&cache=shared", memDBCounter.Add(1))
		database, err = db.New(dbName)
		if err != nil {
			panic(err)
		}
	} else {
		database, err = setupCache(cfg.cacheDir)
		if err != nil {
			panic(err)
		}
	}
	ctx = db.WithDB(ctx, database)

	Logger.Debug("Setup complete")
	return ctx
}

// Cleanup closes the database connection held in the context.
func Cleanup(ctx context.Context) {
	if database := db.DBFromCtx(ctx); database != nil {
		_ = database.Close()
	}
}

// setupCache sets up the cache directory and database, returns the opened DB.
func setupCache(customDir string) (*sql.DB, error) {
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

	err = os.MkdirAll(fp, 0o755)
	if err != nil {
		return nil, err
	}

	return db.New(filepath.Join(fp, "cache.db"))
}

func getHash(ctx context.Context, stepName string, fp string) (h []byte, err error) {
	return db.GetHash(db.DBFromCtx(ctx), stepName, fp)
}

func setHash(ctx context.Context, stepName string, fp string, h []byte) error {
	return db.SetHash(db.DBFromCtx(ctx), stepName, fp, h)
}

func hashFile(ctx context.Context, fp string) (h []byte, err error) {
	hs := HasherFromCtx(ctx)()

	f, err := vfs.FSFromCtx(ctx).Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

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

func needsRebuild(ctx context.Context, stepName string, fp string) (h []byte, err error) {
	h, err = hashFile(ctx, fp)
	if err != nil {
		return nil, err
	}

	hStored, err := getHash(ctx, stepName, fp)
	if err != nil {
		return nil, err
	}

	if hStored == nil || !slices.Equal(hStored, h) {
		return h, nil
	}
	return nil, nil
}
