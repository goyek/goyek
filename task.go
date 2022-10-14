package goyek

// Task represents a named task that can be registered.
// It can consist of a action (function that will be called when task is run),
// dependencies (tasks which has to be run before this one)
// parameters (which can be used within the action).
type Task struct {
	// Name uniquely identifies the task.
	// Names may not be empty and should be easily representable on the CLI.
	Name string

	// Usage provides information what the task does.
	// If it is empty, this task will not be listed in the usage output.
	Usage string

	// Action executes the task in the given flow context.
	// A task can be registered without a action and can act as a "collector" task
	// for a list of dependencies.
	Action func(tf *TF)

	// Deps is a collection of registered tasks that need to be run before this task is executed.
	Deps Deps
}

// RegisteredTask represents a task that has been registered to a Flow.
// It can be used as a dependency for another Task.
type RegisteredTask struct {
	task Task
}

// Name returns the name of the task.
func (r RegisteredTask) Name() string {
	return r.task.Name
}

// Usage returns the description of the task.
func (r RegisteredTask) Usage() string {
	return r.task.Usage
}

// Deps returns the names of all task's dependencies.
func (r RegisteredTask) Deps() []string {
	var deps []string
	for _, task := range r.task.Deps {
		deps = append(deps, task.Name())
	}
	return deps
}

// Deps represents a collection of dependencies.
type Deps []RegisteredTask
