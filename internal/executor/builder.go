package executor

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/db"
	"dario.lol/cf/internal/flags"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

const (
	ansiEraseLine   = "\r\x1b[2K"
	DefaultCacheTTL = 5 * time.Minute
)

type CachedResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type step struct {
	message      string
	silent       bool
	run          func(ctx *Context, progress chan<- string) error
	cacheKey     string
	cacheKeyFunc func(*Context) string
}

type ContextBuilder struct {
	steps           []step
	displayFn       func(ctx *Context)
	invalidatesFunc func(ctx *Context) []string
	skipCache       bool
}

func New() *ContextBuilder {
	return &ContextBuilder{}
}

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

func (b *ContextBuilder) WithZone() *ContextBuilder {
	b.steps = append(b.steps, step{
		message: "Resolving zone",
		run: func(ctx *Context, _ chan<- string) error {
			zoneIdentifier := ctx.Args[0]
			zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
			if err != nil {
				return err
			}
			Set(ctx, ZoneIDKey, zoneID)
			Set(ctx, ZoneNameKey, zoneName)
			return nil
		},
		silent: true,
	})
	return b
}

func (b *ContextBuilder) WithDNSRecord() *ContextBuilder {
	b.steps = append(b.steps, step{
		message: "Resolving DNS record",
		run: func(ctx *Context, _ chan<- string) error {
			zoneID := Get(ctx, ZoneIDKey)
			if zoneID == "" {
				return fmt.Errorf("zone must be resolved before record (call WithZone first)")
			}
			recordIdentifier := ctx.Args[1]
			recordID, recordName, err := cloudflare.LookupDNSRecord(ctx.Client, zoneID, Get(ctx, ZoneNameKey), recordIdentifier)
			if err != nil {
				return err
			}
			Set(ctx, RecordIDKey, recordID)
			Set(ctx, RecordNameKey, recordName)
			return nil
		},
		silent: true,
	})
	return b
}

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

func (b *ContextBuilder) Display(fn func(ctx *Context)) *ContextBuilder {
	b.displayFn = fn
	return b
}

func (b *ContextBuilder) Invalidates(fn func(ctx *Context) []string) *ContextBuilder {
	b.invalidatesFunc = fn
	return b
}

func (b *ContextBuilder) WithConfirmation(message string) *ContextBuilder {
	return b.WithConfirmationFunc(func(ctx *Context) string {
		return message
	})
}

func (b *ContextBuilder) WithConfirmationFunc(fn func(ctx *Context) string) *ContextBuilder {
	b.steps = append(b.steps, step{
		run: func(ctx *Context, _ chan<- string) error {
			if skip, _ := ctx.Cmd.Flags().GetBool(flags.YesFlag); skip {
				return nil
			}
			prompt := fn(ctx)
			confirmed, err := ui.Confirm(prompt)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					return ErrAborted
				}
				return err
			}
			if !confirmed {
				return ErrAborted
			}
			return nil
		},
		silent: true,
	})
	return b
}

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

	for i, s := range b.steps {
		if b.hasCacheKey(s) && b.invalidatesFunc == nil && !b.skipCache && config.Cfg.Caching {
			cacheKey := b.buildCacheKey(ctx, s)
			if b.tryRestoreFromCache(ctx, cacheKey) {
				ctx.Duration = time.Since(start)
				b.displayFn(ctx)
				return
			}
		}

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

		if b.hasCacheKey(s) && ctx.Error == nil && b.invalidatesFunc == nil && config.Cfg.Caching {
			cacheKey := b.buildCacheKey(ctx, s)
			b.storeToCache(ctx, cacheKey, b.steps[i:])
		}
	}

	ctx.Duration = time.Since(start)
	fmt.Fprint(writer, ansiEraseLine)
	_ = writer.Flush()

	if ctx.Error == nil && b.invalidatesFunc != nil {
		tags := b.invalidatesFunc(ctx)
		var exactTags []string
		for _, tag := range tags {
			if strings.HasSuffix(tag, ":") {
				_ = db.InvalidatePrefix(tag)
			} else {
				exactTags = append(exactTags, tag)
			}
		}
		if len(exactTags) > 0 {
			_ = db.InvalidateTags(exactTags)
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

	if ctx.Pagination.Limit > 0 || ctx.Pagination.Page > 1 {
		baseKey = fmt.Sprintf("%s:limit=%d:page=%d", baseKey, ctx.Pagination.Limit, ctx.Pagination.Page)
	}

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

	for _, s := range steps {
		if s.cacheKey != "" {
			_ = db.AddTagsToKey(cacheKey, []string{s.cacheKey})
		}
	}
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
