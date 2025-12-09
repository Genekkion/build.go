package util

import (
	"runtime"
)

func CurrentFilePath() string {
	_, f, _, ok := runtime.Caller(1)
	if !ok {
		panic("no caller found")
	}
	return f
}
