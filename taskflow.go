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
	Verbose     *BoolParam     // when enabled, then the whole output will be always streamed
	DefaultTask RegisteredTask // task which is run when non is explicitly provided

	params map[string]parameter
	tasks  map[string]Task
}

// RegisteredTask represents a task that has been registered to a Taskflow.
// It can be used as a dependency for another Task.
type RegisteredTask struct {
	name string
}

// New return an instance of Taskflow with initialized fields.
func New() *Taskflow {
	return &Taskflow{
		Output: DefaultOutput,
	}
}

// ConfigureValue registers a generic parameter that is defined by the calling code.
// Use this variant in case the primitive-specific implementations cannot cover the parameter.
func (f *Taskflow) ConfigureValue(newValue func() Value, info ParameterInfo) ValueParam {
	f.configure(newValue, info)
	return ValueParam{RegisteredParam{name: info.Name}}
}

// ConfigureBool registers a boolean parameter.
func (f *Taskflow) ConfigureBool(defaultValue bool, info ParameterInfo) BoolParam {
	f.configure(func() Value {
		value := boolValue(defaultValue)
		return &value
	}, info)
	return BoolParam{RegisteredParam{name: info.Name}}
}

// ConfigureInt registers an integer parameter.
func (f *Taskflow) ConfigureInt(defaultValue int, info ParameterInfo) IntParam {
	f.configure(func() Value {
		value := intValue(defaultValue)
		return &value
	}, info)
	return IntParam{RegisteredParam{name: info.Name}}
}

// ConfigureString registers a string parameter.
func (f *Taskflow) ConfigureString(defaultValue string, info ParameterInfo) StringParam {
	f.configure(func() Value {
		value := stringValue(defaultValue)
		return &value
	}, info)
	return StringParam{RegisteredParam{name: info.Name}}
}

func (f *Taskflow) configure(newValue func() Value, info ParameterInfo) {
	if info.Name == "" {
		panic("parameter name cannot be empty")
	}
	if _, exists := f.params[info.Name]; exists {
		panic(fmt.Sprintf("%s parameter was already registered", info.Name))
	}
	if f.params == nil {
		f.params = make(map[string]parameter)
	}
	f.params[info.Name] = parameter{
		info:     info,
		newValue: newValue,
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
