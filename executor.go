package goyek

import (
	"context"
	"io"
)

type executor struct {
	output      io.Writer
	defined     map[string]taskSnapshot
	logger      Logger
	middlewares []Middleware
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (r *executor) Execute(ctx context.Context, tasks []string) error {
	executedTasks := map[string]bool{}
	for _, name := range tasks {
		if err := r.run(ctx, name, executedTasks); err != nil {
			return err
		}
	}
	return nil
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
		return &FailError{Task: task.name}
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

	// apply defined middlewares
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
