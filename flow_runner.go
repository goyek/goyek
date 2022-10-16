package goyek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

type flowRunner struct {
	output       io.Writer
	defined      map[string]taskSnapshot
	logDecorator LogDecorator
	middlewares  []func(Runner) Runner
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

	// prepare default middlewares
	if len(r.middlewares) == 0 {
		r.middlewares = append(r.middlewares, reporter)
	}

	// prepare runner with middlewares
	runner := NewRunner(task.action)
	for _, middleware := range r.middlewares {
		runner = middleware(runner)
	}

	// run action
	in := Input{
		Context:      ctx,
		TaskName:     task.name,
		Output:       r.output,
		LogDecorator: r.logDecorator,
	}
	result := runner(in)
	return result.Status != StatusFailed
}

func reporter(next Runner) Runner {
	// this middleware is copy-pasted from middleware/reporter.go
	// in order to avoid cyclic reference, yet expose all the middlewares under one package
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
		case StatusNotRun, StatusSkipped:
			status = "SKIP"
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
