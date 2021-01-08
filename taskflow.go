/*
Package taskflow helps implementing build automation.
It is intended to be used in concert with the "go run" command,
to run a program which implements the build pipeline (called taskflow).
A taskflow consists of a set of registered tasks.
A task has a name, can have a defined command, which is a function with signature
	func (*taskflow.TF)
and can have dependencies (already defined tasks).

When the taskflow is executed for given tasks,
then the tasks' commands are run in the order defined by their dependencies.
The task's dependencies are run in a recusrive manner, however each is going to be run at most once.

The taskflow is interupted in case a command fails.
Within these functions, use the Error, Fail or related methods to signal failure.
*/
package taskflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// CodePass indicates that taskflow passed.
	CodePass = 0
	// CodeFailure indicates that taskflow failed.
	CodeFailure = 1
	// CodeInvalidArgs indicates that got invalid input.
	CodeInvalidArgs = 2
)

// DefaultOutput is the default output used by Taskflow if it is not set.
var DefaultOutput io.Writer = os.Stdout

// Taskflow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
// By default Taskflow prints to Stdout, but it can be change by setting Output.
type Taskflow struct {
	Output  io.Writer
	Params  Params
	Verbose bool

	tasks map[string]Task
}

// RegisteredTask represents a task that has been registered to a Taskflow.
// It can be used as a dependency for another Task.
type RegisteredTask struct {
	name string
}

// Params represents Taskflow parameters used within Taskflow.
// The default values set in the struct are overridden in Run method.
type Params map[string]string

// New return a valid instance of Taskflow with DefaultOutput and initized Params.
func New() *Taskflow {
	return &Taskflow{
		Output: DefaultOutput,
		Params: Params{},
	}
}

// Register registers the task.
func (f *Taskflow) Register(task Task) (RegisteredTask, error) {
	// validate
	if task.Name == "" {
		return RegisteredTask{}, errors.New("task name cannot be empty")
	}
	if f.isRegistered(task.Name) {
		return RegisteredTask{}, fmt.Errorf("%s task was already registered", task.Name)
	}
	for _, dep := range task.Dependencies {
		if !f.isRegistered(dep.name) {
			return RegisteredTask{}, fmt.Errorf("invalid dependency %s", dep.name)
		}
	}

	f.tasks[task.Name] = task
	return RegisteredTask{name: task.Name}, nil
}

// MustRegister registers the task. It panics in case of any error.
func (f *Taskflow) MustRegister(task Task) RegisteredTask {
	dep, err := f.Register(task)
	if err != nil {
		panic(err)
	}
	return dep
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *Taskflow) Run(ctx context.Context, args ...string) int {
	params := make(Params, len(f.Params))
	for k, v := range f.Params {
		params[k] = v
	}

	flow := &flowRunner{
		output:  f.Output,
		params:  f.Params,
		tasks:   f.tasks,
		verbose: f.Verbose,
	}

	if flow.output == nil {
		flow.output = DefaultOutput
	}

	return flow.Run(ctx, args)
}

func (f *Taskflow) isRegistered(name string) bool {
	if f.tasks == nil {
		f.tasks = map[string]Task{}
	}
	_, ok := f.tasks[name]
	return ok
}
