package db

import (
	"crypto/sha256"
	"testing"

	"github.com/Genekkion/build.go/internal/test"
)

func TestGetSetHash(t *testing.T) {
	t.Parallel()

	db := newTestDb(t)

	fp := "test.txt"
	h := sha256.New().Sum([]byte("test"))

	err := SetHash(db, fp, h)
	test.NilErr(t, err)

	hRes, err := GetHash(db, fp)
	test.NilErr(t, err)

	test.AssertEqual(t, "Expected hash to be equal", h, hRes)
}

func TestGetHash_NotExists(t *testing.T) {
	t.Parallel()

	db := newTestDb(t)

	h, err := GetHash(db, "test.txt")
	test.NilErr(t, err)
	test.AssertEqual(t, "Expected hash to be nil", nil, h)
}

func TestOverwriteHash(t *testing.T) {
	t.Parallel()

	db := newTestDb(t)

	fp := "test.txt"
	h1 := sha256.New().Sum([]byte("test"))
	h2 := sha256.New().Sum([]byte("test2"))

	err := SetHash(db, fp, h1)
	test.NilErr(t, err)

	err = SetHash(db, fp, h2)
	test.NilErr(t, err)

	hRes, err := GetHash(db, fp)
	test.NilErr(t, err)

	test.AssertEqual(t, "Expected hash to be equal", h2, hRes)
}
