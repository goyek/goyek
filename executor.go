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
		defined     map[string]*Task
		middlewares []Middleware
		defaultTask *Task
	}
)

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
//
//nolint:gocyclo // Contains graph traversal logic.
func (r *executor) Execute(in ExecuteInput) error {
	// Handle default task.
	if len(in.Tasks) == 0 && r.defaultTask != nil {
		in.Tasks = append(in.Tasks, r.defaultTask.Name)
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
		if !in.NoDeps && len(task.Deps) > 0 {
			deps := make([]string, 0, len(task.Deps))
			for _, dep := range task.Deps {
				if visited[dep.task.Name] {
					continue
				}
				deps = append(deps, dep.task.Name)
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

		if !task.Parallel {
			// Run task sychronously.
			if err := r.runTask(ctx, task, out, in.Logger); err != nil {
				return err
			}
			continue
		}

		tasksToRun := []*Task{task}

		// Find all parallel tasks that have not beed run
		// and have no dependencies.
		for _, next := range tasks {
			nextTask := r.defined[next]
			if !r.canRunTask(nextTask, visited, in.NoDeps) {
				continue
			}
			// Parallel task has none not-executed dependencies so we can run it.
			visited[nextTask.Name] = true
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
	if len(in.Tasks) == 0 {
		return errors.New("no task provided")
	}
	for _, task := range in.Tasks {
		if task == "" {
			return errors.New("task name cannot be empty")
		}
		if _, ok := r.defined[task]; !ok {
			return errors.New("task provided but not defined: " + task)
		}
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

func (r *executor) canRunTask(task *Task, visited map[string]bool, noDeps bool) bool {
	if visited[task.Name] {
		return false
	}
	if !task.Parallel {
		// We cannot run a non-parallel task in parallel.
		return false
	}

	if noDeps {
		// Dependencies are not honored so we can just run the task.
		return true
	}

	for _, dep := range task.Deps {
		if visited[dep.task.Name] {
			continue
		}
		// The task has a dependency which is not executed yet.
		return false
	}
	return true
}

func (r *executor) runParallelTasks(ctx context.Context, tasks []*Task, output io.Writer, logger Logger) error {
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

func (r *executor) runTask(ctx context.Context, task *Task, output io.Writer, logger Logger) error {
	// prepare runner
	runner := NewRunner(task.Action)

	// apply defined middlewares
	for _, middleware := range r.middlewares {
		runner = middleware(runner)
	}

	// run action
	in := Input{
		Context:  ctx,
		TaskName: task.Name,
		Parallel: task.Parallel,
		Output:   output,
		Logger:   logger,
	}
	result := runner(in)
	if result.Status == StatusFailed {
		return &FailError{Task: task.Name}
	}
	return nil
}
