package taskflow

// Task represents a named task that can be registered.
// It can consist of a command (function that will be called when task is run)
// and dependencies (tasks which has to be run before this one).
type Task struct {
	Name         string
	Description  string
	Command      func(tf *TF)
	Dependencies Deps
	Parameters   []RegisteredParam
}

// Deps represents an collection of registered Tasks.
type Deps []RegisteredTask
