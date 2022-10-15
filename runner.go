package goyek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type runner struct {
	output  io.Writer
	defined map[string]taskSnapshot
	verbose bool
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (r *runner) Run(ctx context.Context, tasks []string) bool {
	from := time.Now()
	executedTasks := map[string]bool{}
	for _, name := range tasks {
		if err := r.run(ctx, name, executedTasks); err != nil {
			fmt.Fprintf(r.output, "%v\t%.3fs\n", err, time.Since(from).Seconds())
			return false
		}
	}
	fmt.Fprintf(r.output, "ok\t%.3fs\n", time.Since(from).Seconds())
	return true
}

func (r *runner) run(ctx context.Context, name string, executed map[string]bool) error {
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

func (r *runner) runTask(ctx context.Context, task taskSnapshot) bool {
	if task.action == nil {
		return true
	}

	writer := r.output
	var streamWriter *strings.Builder
	if !r.verbose {
		streamWriter = &strings.Builder{}
		writer = streamWriter
	}
	writer = &syncWriter{Writer: writer}

	// report task start
	fmt.Fprintf(writer, "===== TASK  %s\n", task.name)

	tf := &TF{
		ctx:    ctx,
		name:   task.name,
		output: writer,
	}
	result := tf.run(task.action)

	// report task end
	status := "PASS"
	passed := true
	switch {
	case result.failed:
		status = "FAIL"
		passed = false
	case result.skipped:
		status = "SKIP"
	}
	fmt.Fprintf(writer, "----- %s: %s (%.2fs)\n", status, task.name, result.duration.Seconds())

	if streamWriter != nil && result.failed {
		io.Copy(r.output, strings.NewReader(streamWriter.String())) //nolint:errcheck,gosec // not checking errors when writing to output
	}

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
