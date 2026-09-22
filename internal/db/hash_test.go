package db

import (
	"crypto/sha256"
	"testing"

	"github.com/Genekkion/build.go/internal/test"
)

func TestGetSetHash(t *testing.T) {
	t.Parallel()

	db := newTestDb(t)

	step := "step1"
	fp := "test.txt"
	h := sha256.New().Sum([]byte("test"))

	err := SetHash(db, step, fp, h)
	test.NilErr(t, err)

	hRes, err := GetHash(db, step, fp)
	test.NilErr(t, err)

	test.AssertEqual(t, "Expected hash to be equal", h, hRes)
}

func TestGetHash_NotExists(t *testing.T) {
	t.Parallel()

	db := newTestDb(t)

	h, err := GetHash(db, "step1", "test.txt")
	test.NilErr(t, err)
	test.AssertEqual(t, "Expected hash to be nil", nil, h)
}

func TestOverwriteHash(t *testing.T) {
	t.Parallel()

	db := newTestDb(t)

	step := "step1"
	fp := "test.txt"
	h1 := sha256.New().Sum([]byte("test"))
	h2 := sha256.New().Sum([]byte("test2"))

	err := SetHash(db, step, fp, h1)
	test.NilErr(t, err)

	err = SetHash(db, step, fp, h2)
	test.NilErr(t, err)

	hRes, err := GetHash(db, step, fp)
	test.NilErr(t, err)

	test.AssertEqual(t, "Expected hash to be equal", h2, hRes)
}

func TestStepScopedIsolation(t *testing.T) {
	t.Parallel()

	db := newTestDb(t)

	fp := "shared.txt"
	h1 := sha256.New().Sum([]byte("step1-hash"))
	h2 := sha256.New().Sum([]byte("step2-hash"))

	err := SetHash(db, "stepA", fp, h1)
	test.NilErr(t, err)

	err = SetHash(db, "stepB", fp, h2)
	test.NilErr(t, err)

	resA, err := GetHash(db, "stepA", fp)
	test.NilErr(t, err)
	test.AssertEqual(t, "stepA hash mismatch", h1, resA)

	resB, err := GetHash(db, "stepB", fp)
	test.NilErr(t, err)
	test.AssertEqual(t, "stepB hash mismatch", h2, resB)
}
