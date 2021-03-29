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
type Taskflow struct {
	Output      io.Writer      // output where text is printed; os.Stdout by default
	Verbose     bool           // when enabled, then the whole output will be always streamed
	DefaultTask RegisteredTask // task which is run when non is explicitly provided

	params map[string]Parameter
	tasks  map[string]Task
}

// Parameter represents the general information of a parameter for one or more tasks.
type Parameter struct {
	Name    string
	Default string
	Usage   string
}

// RegisteredTask represents a task that has been registered to a Taskflow.
// It can be used as a dependency for another Task.
type RegisteredTask struct {
	name string
}

// RegisteredParam represents a parameter that has been registered to a Taskflow.
// It can be used as a parameter for a Task.
type RegisteredParam struct {
	name string
}

// Name returns the key of the parameter.
func (p RegisteredParam) Name() string {
	return p.name
}

// New return an instance of Taskflow with initialized fields.
func New() *Taskflow {
	return &Taskflow{
		Output: DefaultOutput,
	}
}

// Configure adds the given parameter to the set of parameters.
func (f *Taskflow) Configure(param Parameter) (RegisteredParam, error) {
	if param.Name == "" {
		return RegisteredParam{}, errors.New("parameter name cannot be empty")
	}
	if _, exists := f.params[param.Name]; exists {
		return RegisteredParam{}, fmt.Errorf("%s parameter was already registered", param.Name)
	}
	if f.params == nil {
		f.params = make(map[string]Parameter)
	}
	f.params[param.Name] = param
	return RegisteredParam{name: param.Name}, nil
}

// MustConfigure adds the given parameter to the set of parameters. It panics in case of any error.
func (f *Taskflow) MustConfigure(param Parameter) RegisteredParam {
	reg, err := f.Configure(param)
	if err != nil {
		panic(err)
	}
	return reg
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
	flow := &flowRunner{
		output:      f.Output,
		params:      f.params,
		tasks:       f.tasks,
		verbose:     f.Verbose,
		defaultTask: f.DefaultTask,
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
