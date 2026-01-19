package executor

// Key is a typed key for storing/retrieving values from context
type Key[T any] struct {
	name string
}

// NewKey creates a typed key for storing data in context
func NewKey[T any](name string) Key[T] {
	return Key[T]{name: name}
}

// Name returns the key's string name
func (k Key[T]) Name() string {
	return k.name
}
