package goyek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

type executor struct {
	output      io.Writer
	defined     map[string]taskSnapshot
	logger      Logger
	middlewares []func(Runner) Runner
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (r *executor) Execute(ctx context.Context, tasks []string) bool {
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

func (r *executor) run(ctx context.Context, name string, executed map[string]bool) error {
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

func (r *executor) runTask(ctx context.Context, task taskSnapshot) bool {
	// prepare runner
	action := task.action
	if action == nil {
		action = func(tf *TF) {}
	}
	runner := NewRunner(action)

	// apply defined middlewares or the default one
	if len(r.middlewares) == 0 {
		r.middlewares = append(r.middlewares, reporter)
	}
	for _, middleware := range r.middlewares {
		runner = middleware(runner)
	}

	// run action
	in := Input{
		Context:  ctx,
		TaskName: task.name,
		Output:   r.output,
		Logger:   r.logger,
	}
	result := runner(in)
	return result.Status != StatusFailed
}

// reporter is a middleware is copy-pasted from middleware/reporter.go
// It is done so to avoid cyclic reference, yet to expose all the middlewares
// under one package for a more user-friendly API.
func reporter(next Runner) Runner {
	return func(in Input) Result {
		// report start task
		fmt.Fprintf(in.Output, "===== TASK  %s\n", in.TaskName)
		start := time.Now()

		// run
		res := next(in)

		// report task end
		status := "PASS"
		switch res.Status {
		case StatusFailed:
			status = "FAIL"
		case StatusSkipped:
			status = "SKIP"
		case StatusNotRun:
			status = "NOOP"
		}
		fmt.Fprintf(in.Output, "----- %s: %s (%.2fs)\n", status, in.TaskName, time.Since(start).Seconds())

		// report panic if happened
		if res.PanicStack != nil {
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
