package set

import (
	"fmt"
	"slices"
	"testing"

	"greddit/internal/test"
)

func Test_NewSet(t *testing.T) {
	t.Parallel()

	s := New[int]()
	expected := 0
	got := s.Len()

	test.AssertEqual(t, "Expected empty set", expected, got)
}

func Test_NewSetWithSlice(t *testing.T) {
	t.Parallel()

	slice := []int{1, 2, 3, 3}

	// Includes duplicate
	s := New(WithSlice(slice))

	l := s.Len()
	test.AssertEqual(t, "Incorrect length", 3, l)

	if !s.Contains(1) || !s.Contains(2) || !s.Contains(3) {
		expected := []int{1, 2, 3}
		got := []int{}
		for k := range s.m {
			got = append(got, k)
		}
		test.AssertEqual(t, "Unexpected set elements", expected, got)
	}
}

func Test_Add(t *testing.T) {
	t.Parallel()

	s := New[int]()

	modified := s.Add(1, 2, 3)
	test.AssertEqual(t, "Expected set to be modified", true, modified)

	l := s.Len()
	test.AssertEqual(t, "Incorrect set length", 3, l)

	// Add duplicate elements
	modified = s.Add(2, 3)
	test.AssertEqual(t, "Expected set to not be modified", false, modified)

	l = s.Len()
	test.AssertEqual(t, "Incorrect set length", 3, l)
}

func Test_Contains(t *testing.T) {
	t.Parallel()

	s := New[int]()
	s.Add(5, 10)

	test.AssertEqual(t, "Expected set to contain 5", true, s.Contains(5))
	test.AssertEqual(t, "Expected set to contain 10", true, s.Contains(10))
	test.AssertEqual(t, "Expected set not to contain 7", false, s.Contains(7))
}

func Test_Remove(t *testing.T) {
	t.Parallel()

	s := New[int]()
	s.Add(1, 2, 3)

	// Remove existing elements
	modified := s.Remove(2)
	test.AssertEqual(t, "Expected modification when removing existing element", true, modified)
	test.AssertEqual(t, "Expected set not to contain 2 after removal", false, s.Contains(2))
	test.AssertEqual(t, "Expected length 2", 2, s.Len())

	// Remove non-existing element
	modified = s.Remove(42)
	test.AssertEqual(t, "Expected no modification when removing non-existing element", false, modified)
	test.AssertEqual(t, "Expected length 2 after removing non-existing element", 2, s.Len())
}

func Test_Keys(t *testing.T) {
	t.Parallel()

	slice := []int{1, 2, 3}

	s := New(WithSlice(slice))

	keys := s.Keys()

	test.AssertEqual(t, "Incorrect length", len(slice), len(keys))
	for _, v := range slice {
		test.Assert(t, fmt.Sprintf("Expected element in keys: %v", v), slices.Contains(keys, v))
	}
}
