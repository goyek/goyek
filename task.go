package goyek

// Task represents a named task that can be registered.
// It can consist of a command (function that will be called when task is run),
// dependencies (tasks which has to be run before this one)
// parameters (which can be used within the command).
type Task struct {
	Name    string
	Usage   string
	Command func(tf *TF)
	Deps    Deps
	Params  Params
}

// Deps represents a collection of registered Tasks.
type Deps []RegisteredTask

// Params represents a collection of registered Params.
type Params []RegisteredParam
