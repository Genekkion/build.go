package fpath

import (
	"runtime"
)

// CurrentFilePath returns the path to the current file (of the caller).
func CurrentFilePath() string {
	_, f, _, ok := runtime.Caller(1)
	if !ok {
		panic("no caller found")
	}
	return f
}
