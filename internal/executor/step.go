package executor

type StepRunner interface {
	run(ctx *Context, progress chan<- string) error
	getMessage() string
	isSilent() bool
	getCacheKey() string
	getCacheKeyFunc() func(*Context) string
}

type Step[T any] struct {
	key          Key[T]
	message      string
	fn           func(*Context, chan<- string) (T, error)
	silent       bool
	cacheKey     string
	cacheKeyFunc func(*Context) string
}

func NewStep[T any](key Key[T], message string) *Step[T] {
	return &Step[T]{key: key, message: message}
}

func (s *Step[T]) Func(fn func(*Context, chan<- string) (T, error)) *Step[T] {
	s.fn = fn
	return s
}

func (s *Step[T]) Silent() *Step[T] {
	s.silent = true
	return s
}

func (s *Step[T]) CacheKey(key string) *Step[T] {
	s.cacheKey = key
	return s
}

func (s *Step[T]) CacheKeyFunc(fn func(*Context) string) *Step[T] {
	s.cacheKeyFunc = fn
	return s
}

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

func (s *Step[T]) getCacheKey() string {
	return s.cacheKey
}

func (s *Step[T]) getCacheKeyFunc() func(*Context) string {
	return s.cacheKeyFunc
}
