package taskflow

import (
	"context"
	"fmt"
)

// Taskflow TODO.
type Taskflow struct {
	tasks map[string]Task
}

// Dependency TODO.
type Dependency struct {
	name string
}

// Register TODO.
func (f *Taskflow) Register(task Task) (Dependency, error) {
	if f.isRegistered(task.Name) {
		return Dependency{}, fmt.Errorf("%s task was already registered", task.Name) //nolint:goerr113 // TODO
	}
	f.tasks[task.Name] = task
	return Dependency{name: task.Name}, nil
}

// MustRegister TODO.
func (f *Taskflow) MustRegister(task Task) Dependency {
	dep, err := f.Register(task)
	if err != nil {
		panic(err)
	}
	return dep
}

// Execute TODO.
func (f *Taskflow) Execute(ctx context.Context, taskNames ...string) error {
	// validate
	for _, name := range taskNames {
		if !f.isRegistered(name) {
			return fmt.Errorf("%s task was not registered", name) //nolint:goerr113 // TODO
		}
	}

	// run recursive execution
	executedTasks := map[string]bool{}
	for _, name := range taskNames {
		if err := f.execute(ctx, name, executedTasks); err != nil {
			return err
		}
	}
	return nil
}

// MustExecute TODO.
func (f *Taskflow) MustExecute(ctx context.Context, taskNames ...string) {
	if err := f.Execute(ctx, taskNames...); err != nil {
		panic(err)
	}
}

func (f *Taskflow) execute(ctx context.Context, name string, executed map[string]bool) error {
	task := f.tasks[name]
	if executed[name] {
		return nil
	}
	for _, dep := range task.Dependencies {
		if err := f.execute(ctx, dep.name, executed); err != nil {
			return err
		}
	}
	if err := task.Command(&TF{ctx: ctx}); err != nil {
		return fmt.Errorf("%s task failed: %w", name, err) //nolint:goerr113 // TODO
	}
	executed[name] = true
	return nil
}

func (f *Taskflow) isRegistered(name string) bool {
	if f.tasks == nil {
		f.tasks = map[string]Task{}
	}
	_, ok := f.tasks[name]
	return ok
}
