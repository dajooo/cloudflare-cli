package executor

import (
	"encoding/json"
	"time"

	"dario.lol/cf/internal/pagination"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/spf13/cobra"
)

type Context struct {
	Cmd  *cobra.Command
	Args []string

	Client      *cf.Client
	AccountID   string
	Pagination  pagination.Options
	KVNamespace string

	Duration time.Duration
	Error    error

	data map[string]any
}

func newContext(cmd *cobra.Command, args []string) *Context {
	return &Context{
		Cmd:  cmd,
		Args: args,
		data: make(map[string]any),
	}
}

func Set[T any](ctx *Context, key Key[T], value T) {
	ctx.data[key.name] = value
}

func Get[T any](ctx *Context, key Key[T]) T {
	v, ok := ctx.data[key.name]
	if !ok {
		var zero T
		return zero
	}

	if typed, ok := v.(T); ok {
		return typed
	}

	if raw, ok := v.(json.RawMessage); ok {
		var result T
		if err := json.Unmarshal(raw, &result); err == nil {
			ctx.data[key.name] = result
			return result
		}
	}

	var zero T
	return zero
}

func Has[T any](ctx *Context, key Key[T]) bool {
	_, ok := ctx.data[key.name]
	return ok
}
