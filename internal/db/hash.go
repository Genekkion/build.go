package db

import (
	"database/sql"
	"errors"
)

// GetHash returns the hash for the given file path.
func GetHash(db *sql.DB, fp string) (h []byte, err error) {
	const stmt = "SELECT hash FROM hashes WHERE file_path = ?"
	err = db.QueryRow(stmt, fp).Scan(&h)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return h, nil
}

// SetHash sets the hash for the given file path.
func SetHash(db *sql.DB, fp string, h []byte) error {
	const stmt = "INSERT OR REPLACE INTO hashes (file_path, hash) VALUES (?, ?)"
	_, err := db.Exec(stmt, fp, h)
	return err
}
