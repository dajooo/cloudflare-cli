package executor

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/db"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	ansiEraseLine   = "\r\x1b[2K"
	DefaultCacheTTL = 5 * time.Minute
)

// CachedResult stores cached data with timestamp
type CachedResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// step represents a single step in the execution pipeline
type step struct {
	message      string
	silent       bool
	run          func(ctx *Context, progress chan<- string) error
	cacheKey     string                // Static cache key
	cacheKeyFunc func(*Context) string // Dynamic cache key generator
}

// ContextBuilder constructs an executor pipeline with context
type ContextBuilder struct {
	steps           []step
	displayFn       func(ctx *Context)
	invalidatesFunc func(ctx *Context) []string
	skipCache       bool
}

// New creates a new context-based executor builder
func New() *ContextBuilder {
	return &ContextBuilder{}
}

// WithClient adds a step that creates and stores the Cloudflare client
func (b *ContextBuilder) WithClient() *ContextBuilder {
	b.steps = append(b.steps, step{
		message: "Decrypting configuration",
		run: func(ctx *Context, _ chan<- string) error {
			client, err := cloudflare.NewClient()
			if err != nil {
				return err
			}
			ctx.Client = client
			return nil
		},
		silent: true,
	})
	return b
}

// WithAccountID adds a step that resolves and stores the account ID
func (b *ContextBuilder) WithAccountID() *ContextBuilder {
	b.steps = append(b.steps, step{
		message: "Resolving account",
		run: func(ctx *Context, _ chan<- string) error {
			accountID, err := cloudflare.GetAccountID(ctx.Client, ctx.Cmd, "")
			if err != nil {
				return err
			}
			ctx.AccountID = accountID
			return nil
		},
		silent: true,
	})
	return b
}

// WithPagination adds a step that parses pagination flags
func (b *ContextBuilder) WithPagination() *ContextBuilder {
	b.steps = append(b.steps, step{
		run: func(ctx *Context, _ chan<- string) error {
			ctx.Pagination = pagination.GetOptions(ctx.Cmd)
			return nil
		},
		silent: true,
	})
	return b
}

// WithKVNamespace adds a step that resolves and stores the KV namespace ID
func (b *ContextBuilder) WithKVNamespace() *ContextBuilder {
	b.steps = append(b.steps, step{
		run: func(ctx *Context, _ chan<- string) error {
			nsID, _ := ctx.Cmd.Flags().GetString("namespace-id")
			if nsID == "" {
				nsID = config.Cfg.KVNamespaceID
			}
			if nsID == "" {
				return fmt.Errorf("namespace ID is required. Use --namespace-id or 'cf kv namespace switch'")
			}
			ctx.KVNamespace = nsID
			return nil
		},
		silent: true,
	})
	return b
}

// WithNoCache adds a step that reads the --no-cache flag
func (b *ContextBuilder) WithNoCache() *ContextBuilder {
	b.steps = append(b.steps, step{
		run: func(ctx *Context, _ chan<- string) error {
			if noCache, _ := ctx.Cmd.Flags().GetBool("no-cache"); noCache {
				b.skipCache = true
			}
			return nil
		},
		silent: true,
	})
	return b
}

// Step adds a typed step to the pipeline
func (b *ContextBuilder) Step(s StepRunner) *ContextBuilder {
	b.steps = append(b.steps, step{
		message:      s.getMessage(),
		silent:       s.isSilent(),
		run:          s.run,
		cacheKey:     s.getCacheKey(),
		cacheKeyFunc: s.getCacheKeyFunc(),
	})
	return b
}

// Display sets the function to display results
func (b *ContextBuilder) Display(fn func(ctx *Context)) *ContextBuilder {
	b.displayFn = fn
	return b
}

// Invalidates sets the cache invalidation function
func (b *ContextBuilder) Invalidates(fn func(ctx *Context) []string) *ContextBuilder {
	b.invalidatesFunc = fn
	return b
}

// Run returns a cobra run function
func (b *ContextBuilder) Run() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		b.execute(cmd, args)
	}
}

func (b *ContextBuilder) execute(cmd *cobra.Command, args []string) {
	ctx := newContext(cmd, args)
	writer := bufio.NewWriter(os.Stdout)
	fmt.Fprintln(writer)

	start := time.Now()

	// Run all steps, checking cache for steps that have cache keys
	for i, s := range b.steps {
		// Before running a cacheable step, try to restore from cache
		if b.hasCacheKey(s) && b.invalidatesFunc == nil && !b.skipCache {
			cacheKey := b.buildCacheKey(ctx, s)
			if b.tryRestoreFromCache(ctx, cacheKey) {
				// Skip remaining steps except display
				ctx.Duration = time.Since(start)
				b.displayFn(ctx)
				return
			}
		}

		// Run the step
		if s.message != "" && !s.silent {
			err := runStep(writer, s.message, func(progress chan<- string) error {
				return s.run(ctx, progress)
			})
			if err != nil {
				ctx.Error = err
				ctx.Duration = time.Since(start)
				fmt.Fprint(writer, ansiEraseLine)
				_ = writer.Flush()
				b.displayFn(ctx)
				return
			}
			fmt.Fprint(writer, ansiEraseLine)
			_ = writer.Flush()
		} else {
			if err := s.run(ctx, nil); err != nil {
				ctx.Error = err
				ctx.Duration = time.Since(start)
				b.displayFn(ctx)
				return
			}
		}

		// After running a cacheable step, store to cache
		if b.hasCacheKey(s) && ctx.Error == nil && b.invalidatesFunc == nil {
			cacheKey := b.buildCacheKey(ctx, s)
			b.storeToCache(ctx, cacheKey, b.steps[i:])
		}
	}

	ctx.Duration = time.Since(start)
	fmt.Fprint(writer, ansiEraseLine)
	_ = writer.Flush()

	// Handle invalidation
	if ctx.Error == nil && b.invalidatesFunc != nil {
		tags := b.invalidatesFunc(ctx)
		if len(tags) > 0 {
			_ = db.InvalidateTags(tags)
		}
	}

	b.displayFn(ctx)
}

func (b *ContextBuilder) hasCacheKey(s step) bool {
	return s.cacheKey != "" || s.cacheKeyFunc != nil
}

func (b *ContextBuilder) buildCacheKey(ctx *Context, s step) string {
	var baseKey string
	if s.cacheKeyFunc != nil {
		baseKey = s.cacheKeyFunc(ctx)
	} else {
		baseKey = s.cacheKey
	}

	// Append pagination params if enabled
	if ctx.Pagination.Limit > 0 || ctx.Pagination.Page > 1 {
		baseKey = fmt.Sprintf("%s:limit=%d:page=%d", baseKey, ctx.Pagination.Limit, ctx.Pagination.Page)
	}

	// Hash the key
	h := sha256.New()
	h.Write([]byte(baseKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (b *ContextBuilder) tryRestoreFromCache(ctx *Context, cacheKey string) bool {
	cachedBytes, _ := db.Get(db.CacheBucket, []byte(cacheKey))
	if cachedBytes == nil {
		return false
	}

	var cachedResult CachedResult
	if err := json.Unmarshal(cachedBytes, &cachedResult); err != nil {
		return false
	}

	if time.Since(cachedResult.Timestamp) > DefaultCacheTTL {
		return false
	}

	// Store raw JSON - Get[T] will lazily unmarshal to correct type
	var dataMap map[string]json.RawMessage
	if err := json.Unmarshal(cachedResult.Data, &dataMap); err != nil {
		return false
	}

	for k, v := range dataMap {
		ctx.data[k] = v
	}

	return true
}

func (b *ContextBuilder) storeToCache(ctx *Context, cacheKey string, steps []step) {
	dataToCache, err := json.Marshal(ctx.data)
	if err != nil {
		return
	}

	resultToStore := CachedResult{
		Timestamp: time.Now(),
		Data:      dataToCache,
	}

	bytesToStore, err := json.Marshal(resultToStore)
	if err != nil {
		return
	}

	_ = db.Set(db.CacheBucket, []byte(cacheKey), bytesToStore)

	// Use the cache key as the tag for invalidation
	for _, s := range steps {
		if s.cacheKey != "" {
			_ = db.AddTagsToKey(cacheKey, []string{s.cacheKey})
		}
	}
}

func generateCacheKey2(cmd *cobra.Command, args []string) (string, error) {
	var keyParts []string
	keyParts = append(keyParts, cmd.CommandPath())
	keyParts = append(keyParts, args...)

	var flagParts []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Changed {
			flagParts = append(flagParts, fmt.Sprintf("--%s=%s", f.Name, f.Value.String()))
		}
	})
	sort.Strings(flagParts)
	keyParts = append(keyParts, flagParts...)

	h := sha256.New()
	h.Write([]byte(strings.Join(keyParts, ";")))
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func runStep(writer *bufio.Writer, message string, task func(progress chan<- string) error) error {
	s := ui.StyledSpinner()
	resultChan := make(chan error, 1)
	progressChan := make(chan string)
	currentMessage := message

	go func() {
		err := task(progressChan)
		close(progressChan)
		resultChan <- err
	}()

	for {
		select {
		case err := <-resultChan:
			return err
		case msg, ok := <-progressChan:
			if ok {
				fmt.Fprint(writer, ansiEraseLine)
				currentMessage = msg
			}
		default:
			var cmd tea.Cmd
			s, cmd = s.Update(spinner.Tick())
			if cmd != nil {
				_ = cmd()
			}
			fmt.Fprintf(writer, "\r%s %s...", s.View(), currentMessage)
			_ = writer.Flush()
			time.Sleep(50 * time.Millisecond)
		}
	}
}
