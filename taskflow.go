package goyek

import (
	"context"
	"fmt"
	"io"
	"os"
)

const (
	// CodePass indicates that taskflow passed.
	CodePass = 0
	// CodeFailure indicates that taskflow failed.
	CodeFailure = 1
	// CodeInvalidArgs indicates that taskflow got invalid input.
	CodeInvalidArgs = 2
)

// DefaultOutput is the default output used by Taskflow if it is not set.
var DefaultOutput io.Writer = os.Stdout

// Taskflow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
type Taskflow struct {
	Output io.Writer // output where text is printed; os.Stdout by default

	DefaultTask RegisteredTask // task which is run when non is explicitly provided

	verbose *BoolParam // when enabled, then the whole output will be always streamed
	params  map[string]paramValueFactory
	tasks   map[string]Task
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

// VerboseParam returns the out-of-the-box verbose parameter which controls the output behavior.
func (f *Taskflow) VerboseParam() BoolParam {
	if f.verbose == nil {
		param := f.RegisterBoolParam(false, ParamInfo{
			Name:  "v",
			Usage: "Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.",
		})
		f.verbose = &param
	}

	return *f.verbose
}

// RegisterValueParam registers a generic parameter that is defined by the calling code.
// Use this variant in case the primitive-specific implementations cannot cover the parameter.
//
// The value is provided via a factory function since Taskflow could be executed multiple times,
// requiring a new Value instance each time.
func (f *Taskflow) RegisterValueParam(newValue func() ParamValue, info ParamInfo) ValueParam {
	f.registerParam(newValue, info)
	return ValueParam{param{name: info.Name}}
}

// RegisterBoolParam registers a boolean parameter.
func (f *Taskflow) RegisterBoolParam(defaultValue bool, info ParamInfo) BoolParam {
	f.registerParam(func() ParamValue {
		value := boolValue(defaultValue)
		return &value
	}, info)
	return BoolParam{param{name: info.Name}}
}

// RegisterIntParam registers an integer parameter.
func (f *Taskflow) RegisterIntParam(defaultValue int, info ParamInfo) IntParam {
	f.registerParam(func() ParamValue {
		value := intValue(defaultValue)
		return &value
	}, info)
	return IntParam{param{name: info.Name}}
}

// RegisterStringParam registers a string parameter.
func (f *Taskflow) RegisterStringParam(defaultValue string, info ParamInfo) StringParam {
	f.registerParam(func() ParamValue {
		value := stringValue(defaultValue)
		return &value
	}, info)
	return StringParam{param{name: info.Name}}
}

func (f *Taskflow) registerParam(newValue func() ParamValue, info ParamInfo) {
	if info.Name == "" {
		panic("parameter name cannot be empty")
	}
	if _, exists := f.params[info.Name]; exists {
		panic(fmt.Sprintf("%s parameter was already registered", info.Name))
	}
	if f.params == nil {
		f.params = make(map[string]paramValueFactory)
	}
	f.params[info.Name] = paramValueFactory{
		info:     info,
		newValue: newValue,
	}
}

// Register registers the task. It panics in case of any error.
func (f *Taskflow) Register(task Task) RegisteredTask {
	// validate
	if task.Name == "" {
		panic("task name cannot be empty")
	}
	if task.Name[0] == '-' {
		panic("task name cannot start with '-' sign")
	}
	if f.isRegistered(task.Name) {
		panic(fmt.Sprintf("%s task was already registered", task.Name))
	}
	for _, dep := range task.Deps {
		if !f.isRegistered(dep.name) {
			panic(fmt.Sprintf("invalid dependency %s", dep.name))
		}
	}

	f.tasks[task.Name] = task
	return RegisteredTask{name: task.Name}
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *Taskflow) Run(ctx context.Context, args ...string) int {
	// make sure that verbose parameter is registered

	flow := &flowRunner{
		output:      f.Output,
		params:      f.params,
		tasks:       f.tasks,
		verbose:     f.VerboseParam(),
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
