package executor

import (
	"encoding/json"
	"time"

	"dario.lol/cf/internal/pagination"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/spf13/cobra"
)

// Context holds all execution state passed through steps
type Context struct {
	// Command info
	Cmd  *cobra.Command
	Args []string

	// Common values populated by With* methods
	Client      *cf.Client
	AccountID   string
	Pagination  pagination.Options
	KVNamespace string

	// Execution metadata
	Duration time.Duration
	Error    error

	// Typed data store for custom results
	data map[string]any
}

// newContext creates a new context with initialized data map
func newContext(cmd *cobra.Command, args []string) *Context {
	return &Context{
		Cmd:  cmd,
		Args: args,
		data: make(map[string]any),
	}
}

// Set stores a typed value in the context
func Set[T any](ctx *Context, key Key[T], value T) {
	ctx.data[key.name] = value
}

// Get retrieves a typed value from the context.
// If the value was restored from cache (as json.RawMessage), it will be
// unmarshalled to the correct type T on first access.
func Get[T any](ctx *Context, key Key[T]) T {
	v, ok := ctx.data[key.name]
	if !ok {
		var zero T
		return zero
	}

	// If it's already the correct type, return it
	if typed, ok := v.(T); ok {
		return typed
	}

	// If it's raw JSON from cache, unmarshal to T
	if raw, ok := v.(json.RawMessage); ok {
		var result T
		if err := json.Unmarshal(raw, &result); err == nil {
			// Update the cache with the typed value for future calls
			ctx.data[key.name] = result
			return result
		}
	}

	var zero T
	return zero
}

// Has checks if a key exists in the context
func Has[T any](ctx *Context, key Key[T]) bool {
	_, ok := ctx.data[key.name]
	return ok
}
