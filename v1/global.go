package build

import (
	"crypto/sha256"
	"database/sql"
	"io"
	slog2 "log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/Genekkion/build.go/internal/db"
	"github.com/Genekkion/build.go/internal/log/slog"
)

var (
	Logger = func() *slog2.Logger {
		return slog.NewLogger(
			slog.NewHandler(os.Stdout, &slog2.HandlerOptions{
				Level: slog2.LevelInfo,
			}),
		)
	}()
	CacheDir string
	CacheDb  *sql.DB
	Hasher   = sha256.New
)

// Setup sets up the global variables.
// Warning: will panic if unable to set up successfully.
func Setup() {
	var err error

	err = setupCache()
	if err != nil {
		panic(err)
	}

	Logger.Debug("Setup complete",
		"cacheDir", CacheDir,
	)
}

// Cleanup cleans up the global variables.
func Cleanup() {
	CacheDb.Close()
}

// setupCache sets up the cache directory and database.
func setupCache() (err error) {
	fp := filepath.Join(".", ".gobuild")
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

// GetHash returns the hash for the given file path.
func GetHash(fp string) (h []byte, err error) {
	return db.GetHash(CacheDb, fp)
}

// SetHash sets the hash for the given file path.
func SetHash(fp string, h []byte) (err error) {
	return db.SetHash(CacheDb, fp, h)
}

// hashFile returns the hash of the file contents
func hashFile(fp string) (h []byte, err error) {
	hs := Hasher()

	f, err := os.Open(fp)
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
func needsRebuild(fp string) (h []byte, err error) {
	h, err = hashFile(fp)
	if err != nil {
		return nil, err
	}

	hStored, err := GetHash(fp)
	if err != nil {
		return nil, err
	}

	if hStored == nil {
		err = SetHash(fp, h)
		if err != nil {
			return nil, err
		}

	} else if !slices.Equal(hStored, h) {
		return h, nil
	}
	return nil, nil
}
