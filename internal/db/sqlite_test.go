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
