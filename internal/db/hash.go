package db

import (
	"database/sql"
	"errors"
)

// GetHash returns the hash for the given step name and file path.
func GetHash(db *sql.DB, stepName string, fp string) (h []byte, err error) {
	const stmt = "SELECT hash FROM hashes WHERE step_name = ? AND file_path = ?"
	err = db.QueryRow(stmt, stepName, fp).Scan(&h)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return h, nil
}

// SetHash sets the hash for the given step name and file path.
func SetHash(db *sql.DB, stepName string, fp string, h []byte) error {
	const stmt = "INSERT OR REPLACE INTO hashes (step_name, file_path, hash) VALUES (?, ?, ?)"
	_, err := db.Exec(stmt, stepName, fp, h)
	return err
}
