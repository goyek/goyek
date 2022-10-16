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
	logDecorator func(string) string
	interceptors []Interceptor
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
	if len(r.interceptors) == 0 {
		r.interceptors = append(r.interceptors, reporter)
	}

	// prepare runner with interceptors
	taskRunner := taskRunner{task.action}
	runner := taskRunner.run
	for _, interceptor := range r.interceptors {
		runner = interceptor(runner)
	}

	// run action
	in := Input{
		Context:      ctx,
		TaskName:     task.name,
		Output:       r.output,
		LogDecorator: r.logDecorator,
	}
	result := runner(in)
	passed := result.Status != StatusFailed && result.Status != StatusPanicked
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

func reporter(next Runner) Runner {
	// this interceptor is copy-pasted from intercept/reporter.go
	// in order to avoid cyclic reference, yet expose all the interceptors under one package
	return func(in Input) Result {
		// report start task
		fmt.Fprintf(in.Output, "===== TASK  %s\n", in.TaskName)
		start := time.Now()

		// run
		res := next(in)

		// report task end
		status := "PASS"
		switch res.Status {
		case StatusFailed, StatusPanicked:
			status = "FAIL"
		case StatusNotRun, StatusSkipped:
			status = "SKIP"
		}
		fmt.Fprintf(in.Output, "----- %s: %s (%.2fs)\n", status, in.TaskName, time.Since(start).Seconds())

		// report panic if happened
		if res.Status == StatusPanicked {
			if res.PanicValue != nil {
				io.WriteString(in.Output, fmt.Sprintf("panic: %v", res.PanicValue)) //nolint:errcheck,gosec // not checking errors when writing to output
			} else {
				io.WriteString(in.Output, "panic(nil) or runtime.Goexit() called") //nolint:errcheck,gosec // not checking errors when writing to output
			}
			io.WriteString(in.Output, "\n\n") //nolint:errcheck,gosec // not checking errors when writing to output
			in.Output.Write(res.PanicStack)   //nolint:errcheck,gosec // not checking errors when writing to output
		}

		return res
	}
}
