package goyek

import (
	"context"
	"errors"
	"io"
)

type (
	// Executor represents a flow execution function.
	Executor func(ExecuteInput) error

	// ExecuteInput received by the flow executor.
	ExecuteInput struct {
		Context   context.Context
		Tasks     []string
		SkipTasks []string
		NoDeps    bool
		Output    io.Writer
		Logger    Logger
	}

	// ExecutorMiddleware represents a flow execution interceptor.
	ExecutorMiddleware func(Executor) Executor

	executor struct {
		defined     map[string]*taskSnapshot
		middlewares []Middleware
		defaultTask *taskSnapshot
	}
)

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
//
//nolint:gocyclo // Contains graph traversal logic.
func (r *executor) Execute(in ExecuteInput) error {
	// Handle default task.
	if len(in.Tasks) == 0 && r.defaultTask != nil {
		in.Tasks = append(in.Tasks, r.defaultTask.name)
	}

	if err := r.validate(in); err != nil {
		return err
	}

	visited := map[string]bool{}
	for _, skipTask := range in.SkipTasks {
		visited[skipTask] = true
	}

	ctx := in.Context
	tasks := in.Tasks
	out := &syncWriter{Writer: in.Output}
	for len(tasks) > 0 {
		name := tasks[0]
		tasks = tasks[1:]
		task := r.defined[name]
		if visited[name] {
			continue
		}
		if !in.NoDeps && len(task.deps) > 0 {
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
			if err := r.runTask(ctx, task, out, in.Logger); err != nil {
				return err
			}
			continue
		}

		tasksToRun := []*taskSnapshot{task}

		// Find all parallel tasks that have not beed run
		// and have no dependencies.
		for _, next := range tasks {
			nextTask := r.defined[next]
			if !r.canRunTask(nextTask, visited, in.NoDeps) {
				continue
			}
			// Parallel task has none not-executed dependencies so we can run it.
			visited[nextTask.name] = true
			tasksToRun = append(tasksToRun, nextTask)
		}

		// Run parallel tasks.
		if err := r.runParallelTasks(ctx, tasksToRun, out, in.Logger); err != nil {
			return err
		}
	}

	return nil
}

func (r *executor) validate(in ExecuteInput) error {
	for _, task := range in.Tasks {
		if task == "" {
			return errors.New("task name cannot be empty")
		}
		if _, ok := r.defined[task]; !ok {
			return errors.New("task provided but not defined: " + task)
		}
	}

	if len(in.Tasks) == 0 {
		return errors.New("no task provided")
	}

	for _, skippedTask := range in.SkipTasks {
		if skippedTask == "" {
			return errors.New("skipped task name cannot be empty")
		}
		if _, ok := r.defined[skippedTask]; !ok {
			return errors.New("skipped task provided but not defined: " + skippedTask)
		}
	}

	return nil
}

func (r *executor) canRunTask(task *taskSnapshot, visited map[string]bool, noDeps bool) bool {
	if visited[task.name] {
		return false
	}
	if !task.parallel {
		// We cannot run a non-parallel task in parallel.
		return false
	}

	if noDeps {
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

func (r *executor) runParallelTasks(ctx context.Context, tasks []*taskSnapshot, output io.Writer, logger Logger) error {
	var err error
	errCh := make(chan error, len(tasks))
	for _, parallelTask := range tasks {
		parallelTask := parallelTask
		go func() {
			errCh <- r.runTask(ctx, parallelTask, output, logger)
		}()
	}
	for range tasks {
		if runErr := <-errCh; runErr != nil {
			err = runErr
		}
	}
	return err
}

func (r *executor) runTask(ctx context.Context, task *taskSnapshot, output io.Writer, logger Logger) error {
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
		Output:   output,
		Logger:   logger,
	}
	result := runner(in)
	if result.Status == StatusFailed {
		return &FailError{Task: task.name}
	}
	return nil
}
