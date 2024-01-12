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
	tasksToSkip := map[string]bool{}
	for _, skipTask := range skipTasks {
		tasksToSkip[skipTask] = true
	}

	executedTasks := map[string]bool{}
	return r.run(ctx, tasks, executedTasks, tasksToSkip)
}

// run iterativly processes the tasks, but runs them in parallel if possible.
//
//nolint:funlen,gocyclo // This alogorithm is complex.
func (r *executor) run(ctx context.Context, tasks []string, executed, tasksToSkip map[string]bool) error {
	for len(tasks) > 0 {
		name := tasks[0]
		tasks = tasks[1:]
		task := r.defined[name]
		if tasksToSkip[name] {
			continue
		}
		if executed[name] {
			continue
		}
		if !r.noDeps && len(task.deps) > 0 {
			deps := make([]string, 0, len(task.deps))
			for _, dep := range task.deps {
				if tasksToSkip[dep.name] {
					continue
				}
				if executed[dep.name] {
					continue
				}
				deps = append(deps, dep.name)
			}
			if len(deps) > 0 {
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
			if tasksToSkip[other] {
				continue
			}
			if executed[other] {
				continue
			}
			nextTask := r.defined[other]
			if !nextTask.parallel {
				// We cannot run a non-parallel task in parallel.
				continue
			}

			if r.noDeps {
				// Dependencies are not honored so we can just run the task.
				tasksToRun = append(tasksToRun, nextTask)
				continue
			}

			var hasDep bool
			for _, dep := range task.deps {
				if tasksToSkip[dep.name] {
					continue
				}
				if executed[dep.name] {
					continue
				}
				hasDep = true
			}
			if hasDep {
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
		Output:   r.output,
		Logger:   r.logger,
	}
	result := runner(in)
	if result.Status == StatusFailed {
		return &FailError{Task: task.name}
	}
	return nil
}
