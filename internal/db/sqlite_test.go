package db

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/Genekkion/build.go/internal/test"
)

var testCounter = atomic.Int32{}

func newTestDb(t *testing.T) *sql.DB {
	t.Helper()

	count := testCounter.Add(1)
	fp := fmt.Sprintf("test-%d.db", count)
	fp = fmt.Sprintf("file:%s?mode=memory", fp)

	db, err := New(fp)
	test.NilErr(t, err)

	return db
}

func TestMigrationsApplied(t *testing.T) {
	t.Parallel()

	database := newTestDb(t)

	rows, err := database.Query("SELECT version FROM schema_migrations ORDER BY version ASC")
	test.NilErr(t, err)
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		test.NilErr(t, rows.Scan(&v))
		versions = append(versions, v)
	}

	expected := []string{"0000_migrations_table.sql", "0001_step_hashes.sql"}
	test.AssertEqual(t, "Expected all migrations applied", len(expected), len(versions))
	for i := range expected {
		test.AssertEqual(t, "Migration version mismatch", expected[i], versions[i])
	}
}

func TestMigrateIdempotent(t *testing.T) {
	t.Parallel()

	database := newTestDb(t)

	// Second migration run should succeed without error
	err := Migrate(database)
	test.NilErr(t, err)
}
