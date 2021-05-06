package goyek

// Task represents a named task that can be registered.
// It can consist of a command (function that will be called when task is run),
// dependencies (tasks which has to be run before this one)
// parameters (which can be used within the command).
type Task struct {
	// Name uniquely identifies the task.
	// Names may not be empty and should be easily representable on the CLI.
	Name string

	// Usage provides information what the task does.
	// If it is empty, this task will not be listed in the usage output.
	Usage string

	// Command executes the task in the given taskflow context.
	// A task can be registered without a command and can act as a "collector" task
	// for a list of dependencies.
	Command func(tf *TF)

	// Deps lists all registered tasks that need to be run before this task is executed.
	Deps Deps

	// Params is a list of registered parameters that the command may need during executions.
	// Not all parameters need to be queried during execution, yet accessing a parameter
	// that was not registered will fail the task.
	Params Params
}

// Deps represents a collection of registered Tasks.
type Deps []RegisteredTask

// Params represents a collection of registered Params.
type Params []RegisteredParam
