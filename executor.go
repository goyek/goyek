package goyek

import (
	"context"
	"io"
)

type executor struct {
	output      io.Writer
	defined     map[string]*taskSnapshot
	logger      Logger
	middlewares []Middleware
	noDeps      bool
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (r *executor) Execute(ctx context.Context, tasks []string, skipTasks []string) error {
	executed := map[string]bool{}
	for _, skipTask := range skipTasks {
		executed[skipTask] = true
	}

	for len(tasks) > 0 {
		name := tasks[0]
		tasks = tasks[1:]
		task := r.defined[name]
		if executed[name] {
			continue
		}
		if !r.noDeps && len(task.deps) > 0 {
			deps := make([]string, 0, len(task.deps))
			for _, dep := range task.deps {
				if executed[dep.name] {
					continue
				}
				deps = append(deps, dep.name)
			}
			if len(deps) > 0 {
				// Add dependencies to be run first.
				deps = append(deps, name)
				tasks = append(deps, tasks...)
				continue
			}
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if !task.parallel {
			// Run task sychronously.
			if err := r.runTask(ctx, task); err != nil {
				return err
			}
			executed[name] = true
			continue
		}

		tasksToRun := []*taskSnapshot{task}

		// Find all parallel tasks that have not beed run
		// and have no dependencies.
		for _, other := range tasks {
			nextTask := r.defined[other]
			if !r.canRunTask(nextTask, executed) {
				continue
			}
			// Parallel task has no not-executed dependencies so we can run it.
			tasksToRun = append(tasksToRun, nextTask)
		}

		// Run parallel tasks.
		var err error
		errCh := make(chan error, len(tasksToRun))
		for _, parallelTask := range tasksToRun {
			parallelTask := parallelTask
			go func() {
				errCh <- r.runTask(ctx, parallelTask)
			}()
			executed[parallelTask.name] = true
		}
		for range tasksToRun {
			if runErr := <-errCh; runErr != nil {
				err = runErr
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *executor) canRunTask(task *taskSnapshot, executed map[string]bool) bool {
	if executed[task.name] {
		return false
	}
	if !task.parallel {
		// We cannot run a non-parallel task in parallel.
		return false
	}

	if r.noDeps {
		// Dependencies are not honored so we can just run the task.
		return true
	}

	for _, dep := range task.deps {
		if executed[dep.name] {
			continue
		}
		// The task has a not executed dependency.
		return false
	}
	return true
}

func (r *executor) runTask(ctx context.Context, task *taskSnapshot) error {
	// prepare runner
	runner := NewRunner(task.action)

	// apply defined middlewares
	for _, middleware := range r.middlewares {
		runner = middleware(runner)
	}

	// run action
	in := Input{
		Context:  ctx,
		TaskName: task.name,
		Parallel: task.parallel,
		Output:   r.output,
		Logger:   r.logger,
	}
	result := runner(in)
	if result.Status == StatusFailed {
		return &FailError{Task: task.name}
	}
	return nil
}
