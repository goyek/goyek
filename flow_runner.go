package goyek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

type flowRunner struct {
	output       io.Writer
	defined      map[string]taskSnapshot
	verbose      bool
	logDecorator func(string) string
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (r *flowRunner) Run(ctx context.Context, tasks []string) bool {
	from := time.Now()
	executedTasks := map[string]bool{}
	for _, name := range tasks {
		if err := r.run(ctx, name, executedTasks); err != nil {
			fmt.Fprintf(r.output, "%v\t%.3fs\n", err, time.Since(from).Seconds()) // TODO: move to Main
			return false
		}
	}
	fmt.Fprintf(r.output, "ok\t%.3fs\n", time.Since(from).Seconds()) // TODO: move to Main
	return true
}

func (r *flowRunner) run(ctx context.Context, name string, executed map[string]bool) error {
	task := r.defined[name]
	if executed[name] {
		return nil
	}
	for _, dep := range task.deps {
		if err := r.run(ctx, dep, executed); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !r.runTask(ctx, task) {
		return errors.New("task failed")
	}
	executed[name] = true
	return nil
}

func (r *flowRunner) runTask(ctx context.Context, task taskSnapshot) bool {
	if task.action == nil {
		return true
	}

	// prepare default interceptors
	var interceptors []interceptor
	interceptors = append(interceptors, reporter)
	if !r.verbose {
		interceptors = append(interceptors, silentNonFailed)
	}

	// prepare runner with interceptors
	taskRunner := taskRunner{task.action}
	runner := taskRunner.run
	for _, interceptor := range interceptors {
		runner = interceptor(runner)
	}

	// run action
	in := input{
		Context:      ctx,
		TaskName:     task.name,
		Output:       r.output,
		LogDecorator: r.logDecorator,
	}
	result := runner(in)
	passed := result.status != statusFailed && result.status != statusPanicked
	return passed
}

type syncWriter struct {
	io.Writer
	mtx sync.Mutex
}

func (w *syncWriter) Write(p []byte) (int, error) {
	defer func() { w.mtx.Unlock() }()
	w.mtx.Lock()
	return w.Writer.Write(p)
}
