package db

import (
	"database/sql"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

var (
	//go:embed schema.sql
	schemaSQL string
)

// New creates a new sqlite database at the given path.
func New(fp string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", fp)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(schemaSQL)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return db, nil
}
