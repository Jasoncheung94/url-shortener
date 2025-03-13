package ptr

// Of returns a pointer to the given value.
// It’s a generic utility useful for literals or inline values.
// Example: ptr.Of("hello")
func Of[T any](v T) *T {
	return &v
}
