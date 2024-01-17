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
	visited := map[string]bool{}
	for _, skipTask := range skipTasks {
		visited[skipTask] = true
	}

	for len(tasks) > 0 {
		name := tasks[0]
		tasks = tasks[1:]
		task := r.defined[name]
		if visited[name] {
			continue
		}
		if !r.noDeps && len(task.deps) > 0 {
			deps := make([]string, 0, len(task.deps))
			for _, dep := range task.deps {
				if visited[dep.name] {
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

		visited[name] = true

		if !task.parallel {
			// Run task sychronously.
			if err := r.runTask(ctx, task); err != nil {
				return err
			}
			continue
		}

		tasksToRun := []*taskSnapshot{task}

		// Find all parallel tasks that have not beed run
		// and have no dependencies.
		for _, other := range tasks {
			nextTask := r.defined[other]
			if !r.canRunTask(nextTask, visited) {
				continue
			}
			// Parallel task has none not-executed dependencies so we can run it.
			visited[nextTask.name] = true
			tasksToRun = append(tasksToRun, nextTask)
		}

		// Run parallel tasks.
		if err := r.runParallelTasks(ctx, tasksToRun); err != nil {
			return err
		}
	}

	return nil
}

func (r *executor) canRunTask(task *taskSnapshot, visited map[string]bool) bool {
	if visited[task.name] {
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
		if visited[dep.name] {
			continue
		}
		// The task has a dependency which is not executed yet.
		return false
	}
	return true
}

func (r *executor) runParallelTasks(ctx context.Context, tasks []*taskSnapshot) error {
	var err error
	errCh := make(chan error, len(tasks))
	for _, parallelTask := range tasks {
		parallelTask := parallelTask
		go func() {
			errCh <- r.runTask(ctx, parallelTask)
		}()
	}
	for range tasks {
		if runErr := <-errCh; runErr != nil {
			err = runErr
		}
	}
	return err
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
