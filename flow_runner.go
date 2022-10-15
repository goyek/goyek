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

	writer := r.output
	var streamWriter *strings.Builder
	if !r.verbose {
		streamWriter = &strings.Builder{}
		writer = streamWriter
	}
	writer = &syncWriter{Writer: writer}

	// report task start
	fmt.Fprintf(writer, "===== TASK  %s\n", task.name)
	start := time.Now()

	// run task
	tf := &TF{
		ctx:       ctx,
		name:      task.name,
		output:    writer,
		decorator: r.logDecorator,
	}
	result := tf.run(task.action)

	// report task end
	status := "PASS"
	passed := true
	switch result.status {
	case statusFailed, statusPanicked:
		status = "FAIL"
		passed = false
	case statusNotRun, statusSkipped:
		status = "SKIP"
	}
	fmt.Fprintf(writer, "----- %s: %s (%.2fs)\n", status, task.name, time.Since(start).Seconds())

	// report panic if happened
	if result.status == statusPanicked {
		if result.panicValue != nil {
			io.WriteString(tf.output, fmt.Sprintf("panic: %v", result.panicValue)) //nolint:errcheck,gosec // not checking errors when writing to output
		} else {
			io.WriteString(tf.output, "panic(nil) or runtime.Goexit() called") //nolint:errcheck,gosec // not checking errors when writing to output
		}
		io.WriteString(tf.output, "\n\n")  //nolint:errcheck,gosec // not checking errors when writing to output
		tf.output.Write(result.panicStack) //nolint:errcheck,gosec // not checking errors when writing to output
	}

	if streamWriter != nil && !passed {
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
