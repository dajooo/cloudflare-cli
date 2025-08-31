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

	"dario.lol/cf/internal/db"
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

type CachedResult struct {
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type Executor[S any, T any] struct {
	cmd             *cobra.Command
	args            []string
	setupMessage    string
	setup           func() (S, error)
	fetchingMessage string
	fetch           func(setupResult S, cmd *cobra.Command, args []string, progress chan<- string) (T, error)
	display         func(data T, fetchDuration time.Duration, err error)
	cachesFunc      func(cmd *cobra.Command, args []string) ([]string, error)
	invalidatesFunc func(cmd *cobra.Command, args []string, result T) []string
	skipCache       bool
}

type Builder[S any, T any] struct {
	executor *Executor[S, T]
}

func NewBuilder[S any, T any]() *Builder[S, T] {
	return &Builder[S, T]{executor: &Executor[S, T]{}}
}

func (b *Builder[S, T]) Setup(message string, task func() (S, error)) *Builder[S, T] {
	b.executor.setupMessage = message
	b.executor.setup = task
	return b
}

func (b *Builder[S, T]) Fetch(message string, task func(S, *cobra.Command, []string, chan<- string) (T, error)) *Builder[S, T] {
	b.executor.fetchingMessage = message
	b.executor.fetch = task
	return b
}

func (b *Builder[S, T]) Display(displayFunc func(T, time.Duration, error)) *Builder[S, T] {
	b.executor.display = displayFunc
	return b
}

func (b *Builder[S, T]) Caches(task func(cmd *cobra.Command, args []string) ([]string, error)) *Builder[S, T] {
	b.executor.cachesFunc = task
	return b
}

func (b *Builder[S, T]) Invalidates(task func(cmd *cobra.Command, args []string, result T) []string) *Builder[S, T] {
	b.executor.invalidatesFunc = task
	return b
}

func (b *Builder[S, T]) SkipCache(skip bool) *Builder[S, T] {
	b.executor.skipCache = skip
	return b
}

func (b *Builder[S, T]) Build() *Executor[S, T] {
	if b.executor.fetch == nil || b.executor.display == nil {
		panic("Executor is not fully configured: Fetch and Display are required.")
	}
	return b.executor
}

func (e *Executor[S, T]) CobraRun() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		e.cmd = cmd
		e.args = args
		e.Execute()
	}
}

func (e *Executor[S, T]) Execute() {
	var zeroT T
	writer := bufio.NewWriter(os.Stdout)
	fmt.Fprintln(writer)

	if e.cachesFunc != nil && e.invalidatesFunc == nil && !e.skipCache {
		cacheKey, err := generateCacheKey(e.cmd, e.args)
		if err == nil {
			cachedBytes, _ := db.Get(db.CacheBucket, []byte(cacheKey))
			if cachedBytes != nil {
				var cachedResult CachedResult
				if err := json.Unmarshal(cachedBytes, &cachedResult); err == nil {
					if time.Since(cachedResult.Timestamp) <= DefaultCacheTTL {
						var result T
						if err := json.Unmarshal(cachedResult.Data, &result); err == nil {
							e.display(result, 0, nil)
							return
						}
					}
				}
			}
		}
	}

	var setupResult S
	var setupErr error
	if e.setup != nil {
		setupResult, _, setupErr = runStage(writer, e.setupMessage, func() (S, error) { return e.setup() })
		if setupErr != nil {
			e.display(zeroT, 0, setupErr)
			return
		}
		fmt.Fprint(writer, ansiEraseLine)
		_ = writer.Flush()
	}

	fetchResult, fetchDuration, fetchErr := runStageWithProgress(writer, e.fetchingMessage, func(p chan<- string) (T, error) {
		return e.fetch(setupResult, e.cmd, e.args, p)
	})

	fmt.Fprint(writer, ansiEraseLine)
	_ = writer.Flush()

	if fetchErr == nil {
		if e.invalidatesFunc != nil {
			tagsToClear := e.invalidatesFunc(e.cmd, e.args, fetchResult)
			if len(tagsToClear) > 0 {
				_ = db.InvalidateTags(tagsToClear)
			}
		} else if e.cachesFunc != nil {
			cacheKey, err := generateCacheKey(e.cmd, e.args)
			if err == nil {
				tags, err := e.cachesFunc(e.cmd, e.args)
				if err == nil && len(tags) > 0 {
					dataToCache, err := json.Marshal(fetchResult)
					if err == nil {
						resultToStore := CachedResult{
							Timestamp: time.Now(),
							Data:      dataToCache,
						}
						bytesToStore, err := json.Marshal(resultToStore)
						if err == nil {
							_ = db.Set(db.CacheBucket, []byte(cacheKey), bytesToStore)
							_ = db.AddTagsToKey(cacheKey, tags)
						}
					}
				}
			}
		}
	}

	e.display(fetchResult, fetchDuration, fetchErr)
}

func generateCacheKey(cmd *cobra.Command, args []string) (string, error) {
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

type result[T any] struct {
	res      T
	err      error
	duration time.Duration
}

func runStage[T any](writer *bufio.Writer, message string, task func() (T, error)) (T, time.Duration, error) {
	s := ui.StyledSpinner()
	resultChan := make(chan result[T], 1)

	go func() {
		start := time.Now()
		res, err := task()
		duration := time.Since(start)
		resultChan <- result[T]{res: res, err: err, duration: duration}
	}()

	for {
		select {
		case res := <-resultChan:
			return res.res, res.duration, res.err
		default:
			var cmd tea.Cmd
			s, cmd = s.Update(spinner.Tick())
			if cmd != nil {
				_ = cmd()
			}
			fmt.Fprintf(writer, "\r%s %s...", s.View(), message)
			_ = writer.Flush()
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func runStageWithProgress[T any](writer *bufio.Writer, initialMessage string, task func(progress chan<- string) (T, error)) (T, time.Duration, error) {
	s := ui.StyledSpinner()
	resultChan := make(chan result[T], 1)
	progressChan := make(chan string)
	currentMessage := initialMessage

	go func() {
		start := time.Now()
		res, err := task(progressChan)
		duration := time.Since(start)
		close(progressChan)
		resultChan <- result[T]{res: res, err: err, duration: duration}
	}()

	for {
		select {
		case res := <-resultChan:
			return res.res, res.duration, res.err
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
