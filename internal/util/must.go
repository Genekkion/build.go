package util

// Must panics if the given function returns an error. Only to be used for
// testing purposes, strictly not for production.
func Must[T any](f func() (T, error)) T {
	v, err := f()
	if err != nil {
		panic(err)
	}
	return v
}
