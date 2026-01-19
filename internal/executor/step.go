package executor

// StepRunner is the interface that typed steps implement
type StepRunner interface {
	run(ctx *Context, progress chan<- string) error
	getMessage() string
	isSilent() bool
}

// Step is a typed step builder
type Step[T any] struct {
	key     Key[T]
	message string
	fn      func(*Context, chan<- string) (T, error)
	silent  bool
}

// NewStep creates a new typed step with a key and message
func NewStep[T any](key Key[T], message string) *Step[T] {
	return &Step[T]{key: key, message: message}
}

// Func sets the function to run for this step
func (s *Step[T]) Func(fn func(*Context, chan<- string) (T, error)) *Step[T] {
	s.fn = fn
	return s
}

// Silent marks this step to run without a spinner
func (s *Step[T]) Silent() *Step[T] {
	s.silent = true
	return s
}

// StepRunner interface implementation

func (s *Step[T]) run(ctx *Context, progress chan<- string) error {
	result, err := s.fn(ctx, progress)
	if err != nil {
		return err
	}
	Set(ctx, s.key, result)
	return nil
}

func (s *Step[T]) getMessage() string {
	return s.message
}

func (s *Step[T]) isSilent() bool {
	return s.silent
}
