package executor

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"dario.lol/cf/internal/ui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const (
	ansiEraseLine = "\r\x1b[2K"
)

type Executor[S any, T any] struct {
	cmd  *cobra.Command
	args []string

	setupMessage    string
	setup           func() (S, error)
	fetchingMessage string
	fetch           func(setupResult S, cmd *cobra.Command, args []string, progress chan<- string) (T, error)
	display         func(data T, fetchDuration time.Duration, err error)
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

	var setupResult S
	var setupErr error
	if e.setup != nil {
		setupTask := func() (S, error) { return e.setup() }
		setupResult, _, setupErr = runStage(writer, e.setupMessage, setupTask)
		if setupErr != nil {
			fmt.Fprint(writer, ansiEraseLine)
			_ = writer.Flush()
			e.display(zeroT, 0, setupErr)
			return
		}
		fmt.Fprint(writer, ansiEraseLine)
		_ = writer.Flush()
	}

	fetchTask := func(progress chan<- string) (T, error) {
		return e.fetch(setupResult, e.cmd, e.args, progress)
	}
	fetchResult, fetchDuration, fetchErr := runStageWithProgress(writer, e.fetchingMessage, fetchTask)

	fmt.Fprint(writer, ansiEraseLine)
	_ = writer.Flush()

	e.display(fetchResult, fetchDuration, fetchErr)
}

type Builder[S any, T any] struct {
	executor *Executor[S, T]
}

func NewBuilder[S any, T any]() *Builder[S, T] {
	return &Builder[S, T]{
		executor: &Executor[S, T]{},
	}
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

func (b *Builder[S, T]) Build() *Executor[S, T] {
	if b.executor.fetch == nil || b.executor.display == nil {
		// We panic here because this is a developer error (misconfiguration), not a runtime error.
		panic("Executor is not fully configured: Fetch and Display are required.")
	}
	return b.executor
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
		close(progressChan) // Signal that no more progress messages will be sent
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
