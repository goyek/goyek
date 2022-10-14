package goyek

// Task represents a named task that can be registered.
// It can consist of a action (function that will be called when task is run),
// dependencies (tasks which has to be run before this one)
// parameters (which can be used within the action).
type Task struct {
	// Name uniquely identifies the task.
	// Names cannot  be empty and should be easily representable on the CLI.
	Name string

	// Usage provides information what the task does.
	Usage string

	// Action executes the task in the given flow context.
	// A task can be registered without a action and can act as a "collector" task
	// for a list of dependencies (also called "pipeline").
	Action func(tf *TF)

	// Deps is a collection of registered tasks that need to be run before this task is executed.
	Deps Deps
}

// RegisteredTask represents a task that has been registered.
// It can be used as a dependency for another task.
type RegisteredTask struct {
	taskSnapshot
}

// Name returns the name of the task.
func (r RegisteredTask) Name() string {
	return r.name
}

// Usage returns the description of the task.
func (r RegisteredTask) Usage() string {
	return r.usage
}

// Deps returns the names of all task's dependencies.
func (r RegisteredTask) Deps() []string {
	count := len(r.deps)
	if count == 0 {
		return nil
	}
	deps := make([]string, count)
	copy(deps, r.deps)
	return deps
}

// Deps represents a collection of dependencies.
type Deps []RegisteredTask
