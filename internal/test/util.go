package test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// printCaller prints the caller of the function for debugging purposes.
func printCaller(jumps ...int) {
	jump := 2
	if len(jumps) > 0 {
		jump += jumps[0]
	}
	pc, file, line, ok := runtime.Caller(jump)
	if ok {
		fn := runtime.FuncForPC(pc)
		fmt.Printf("Called from %s:%d (%s)\n",
			filepath.Base(file), line, fn.Name())
	}
}

// Assert prints the error message out in a standard format.
func Assert(t *testing.T, message string, got bool, jumps ...int) {
	t.Helper()
	if got {
		return
	}

	t.Errorf("%s", message)
	fmt.Printf("Expression is false\n")

	printCaller(jumps...)
	t.FailNow()
}

// AssertEqual checks if two values are equal using reflect.DeepEqual and fails if not.
func AssertEqual[T any](t *testing.T, message string, expected T, got T, jumps ...int) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("%s", message)
		fmt.Printf("Expected: %v\n", expected)
		fmt.Printf("Got     : %v\n", got)

		printCaller(jumps...)
		t.FailNow()
	}
}

// AssertJsonEqual compares two JSON objects and fails the test if they are not equal.
func AssertJsonEqual(t *testing.T, message string, expected any, got any) {
	t.Helper()
	b1, err := json.Marshal(expected)
	NilErr(t, err)
	b2, err := json.Marshal(got)
	NilErr(t, err)
	AssertEqual(t, message, string(b1), string(b2))
}

// NilErr checks if err is nil, and if not, fails the test.
func NilErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}

	t.Errorf("Expected nil error")
	fmt.Printf("Error: %v\n", err)

	printCaller()
	t.FailNow()
}
