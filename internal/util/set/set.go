package set

// For use with the underlying map value.
var emptyValue = struct{}{}

// Set is the ype for the generic set type.
type Set[T comparable] struct {
	m map[T]struct{}
}

// New creates a new set, with the options specified.
func New[T comparable](opts ...Option[T]) Set[T] {
	s := Set[T]{
		m: make(map[T]struct{}),
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s
}

// Len Returns the number of elements in the set.
func (s Set[T]) Len() int {
	return len(s.m)
}

// Contains Returns whether the set contains the specified key.
func (s Set[T]) Contains(key T) bool {
	_, ok := s.m[key]
	return ok
}

// Add adds the keys to the set. Returns whether the set has been modified.
func (s *Set[T]) Add(keys ...T) (modified bool) {
	modified = false

	for _, key := range keys {
		v := s.Contains(key)
		if v {
			continue
		}

		s.m[key] = emptyValue
		modified = true
	}

	return modified
}

// Remove removes the keys specified from the set. Returns whether the set has
// been modified.
func (s *Set[T]) Remove(keys ...T) (modified bool) {
	modified = false

	for _, key := range keys {
		v := s.Contains(key)
		if !v {
			continue
		}

		delete(s.m, key)
		modified = true
	}
	return modified
}

// Keys returns the keys in the set.
func (s *Set[T]) Keys() []T {
	keys := make([]T, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}

	return keys
}
