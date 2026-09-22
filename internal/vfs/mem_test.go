package vfs_test

import (
	"io"
	"os"
	"slices"
	"testing"

	"github.com/Genekkion/build.go/internal/vfs"
)

func TestMemFS(t *testing.T) {
	mem := vfs.NewMemFS()

	err := mem.WriteFile("/app/main.go", []byte("package main"), 0o644)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	err = mem.WriteFile("/app/util.go", []byte("package main"), 0o644)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	// Test Open
	rc, err := mem.Open("/app/main.go")
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	if string(data) != "package main" {
		t.Fatalf("unexpected content: %s", string(data))
	}

	// Test Stat
	fi, err := mem.Stat("/app/main.go")
	if err != nil {
		t.Fatalf("failed to stat: %v", err)
	}
	if fi.Name() != "main.go" || fi.Size() != int64(len("package main")) {
		t.Fatalf("unexpected file info: %+v", fi)
	}

	// Test Glob exact
	matches, err := mem.Glob("/app/main.go")
	if err != nil || len(matches) != 1 || matches[0] != "/app/main.go" {
		t.Fatalf("unexpected exact glob: %v, %v", matches, err)
	}

	// Test Glob wildcard
	matches, err = mem.Glob("/app/*.go")
	if err != nil || len(matches) != 2 {
		t.Fatalf("unexpected wildcard glob: %v, %v", matches, err)
	}
	if !slices.Equal(matches, []string{"/app/main.go", "/app/util.go"}) {
		t.Fatalf("matches not sorted or incorrect: %v", matches)
	}

	// Test Remove
	if err := mem.Remove("/app/util.go"); err != nil {
		t.Fatalf("failed to remove: %v", err)
	}
	_, err = mem.Open("/app/util.go")
	if err != os.ErrNotExist {
		t.Fatalf("expected ErrNotExist, got %v", err)
	}
}
