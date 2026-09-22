package fpath_test

import (
	"strings"
	"testing"

	"github.com/Genekkion/build.go/v1/fpath"
)

func TestCurrentFilePath(t *testing.T) {
	p := fpath.CurrentFilePath()
	if !strings.HasSuffix(p, "path_test.go") {
		t.Fatalf("expected current file to end with path_test.go, got %s", p)
	}
}
