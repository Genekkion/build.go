package test

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Assert checks if condition is true and fails the test if not.
func Assert(t *testing.T, message string, got bool) {
	t.Helper()
	if !got {
		t.Fatalf("%s: expression is false", message)
	}
}

// AssertEqual checks if two values are equal using reflect.DeepEqual and fails if not.
func AssertEqual[T any](t *testing.T, message string, expected T, got T) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("%s\nExpected: %v\nGot     : %v", message, expected, got)
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
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}
