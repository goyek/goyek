package goyek

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
)

const (
	// CodePass indicates that flow passed.
	CodePass = 0
	// CodeFail indicates that flow failed.
	CodeFail = 1
	// CodeInvalidArgs indicates that flow got invalid input.
	CodeInvalidArgs = 2
)

// Flow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
type Flow struct {
	Output io.Writer // output where text is printed; os.Stdout by default

	DefaultTask RegisteredTask // task which is run when non is explicitly provided

	verbose *RegisteredBoolParam   // when enabled, then the whole output will be always streamed
	workDir *RegisteredStringParam // sets the working directory
	params  map[string]registeredParam
	tasks   map[string]Task
}

// Tasks returns all registered tasks.
func (f *Flow) Tasks() []RegisteredTask {
	var tasks []RegisteredTask
	for _, task := range f.tasks {
		tasks = append(tasks, RegisteredTask{
			task: task,
		})
	}
	return tasks
}

// Params returns all registered parameters.
func (f *Flow) Params() []RegisteredParam {
	// make sure OOTB parameters are registered
	f.VerboseParam()
	f.WorkDirParam()

	var params []RegisteredParam
	for _, param := range f.params {
		params = append(params, param)
	}
	return params
}

// VerboseParam returns the out-of-the-box verbose parameter which controls the output behavior.
// This can be overridden with RegisterVerboseParam.
func (f *Flow) VerboseParam() RegisteredBoolParam {
	if f.verbose == nil {
		param := f.RegisterBoolParam(BoolParam{
			Name:  "v",
			Usage: "Verbose: log all tasks as they are run.",
		})
		f.verbose = &param
	}

	return *f.verbose
}

// RegisterVerboseParam overwrites the default name, usage and value for the verbose parameter.
// If this function is used, the default 'v' parameter will be replaced with this parameter.
func (f *Flow) RegisterVerboseParam(p BoolParam) RegisteredBoolParam {
	param := f.RegisterBoolParam(p)
	f.verbose = &param
	return *f.verbose
}

// WorkDirParam returns the out-of-the-box working directory parameter which controls the working directory.
func (f *Flow) WorkDirParam() RegisteredStringParam {
	if f.workDir == nil {
		param := f.RegisterStringParam(StringParam{
			Name:    "wd",
			Usage:   "Working directory: set the working directory.",
			Default: ".",
		})
		f.workDir = &param
	}

	return *f.workDir
}

// RegisterWorkDirParam overwrites the default name, usage and value for the work dir parameter.
// If this function is used, the default 'wd' parameter will be replaced with this parameter.
func (f *Flow) RegisterWorkDirParam(p StringParam) RegisteredStringParam {
	param := f.RegisterStringParam(p)
	f.workDir = &param
	return *f.workDir
}

// RegisterValueParam registers a generic parameter that is defined by the calling code.
// Use this variant in case the primitive-specific implementations cannot cover the parameter.
//
// The value is provided via a factory function since flow could be executed multiple times,
// requiring a new Value instance each time.
func (f *Flow) RegisterValueParam(p ValueParam) RegisteredValueParam {
	regParam := registeredParam{
		name:     p.Name,
		usage:    p.Usage,
		newValue: p.NewValue,
		required: p.Required,
	}
	f.registerParam(regParam)
	return RegisteredValueParam{regParam}
}

// RegisterBoolParam registers a boolean parameter.
func (f *Flow) RegisterBoolParam(p BoolParam) RegisteredBoolParam {
	valGetter := func() ParamValue {
		return &boolValue{value: p.Default}
	}
	f.registerParam(registeredParam{
		name:     p.Name,
		usage:    p.Usage,
		newValue: valGetter,
		required: p.Required,
	})
	return RegisteredBoolParam{registeredParam{name: p.Name}}
}

// RegisterIntParam registers an integer parameter.
func (f *Flow) RegisterIntParam(p IntParam) RegisteredIntParam {
	valGetter := func() ParamValue {
		return &intValue{value: p.Default}
	}
	regParam := registeredParam{
		name:     p.Name,
		usage:    p.Usage,
		newValue: valGetter,
		required: p.Required,
	}
	f.registerParam(regParam)
	return RegisteredIntParam{regParam}
}

// RegisterStringParam registers a string parameter.
func (f *Flow) RegisterStringParam(p StringParam) RegisteredStringParam {
	valGetter := func() ParamValue {
		return &stringValue{value: p.Default}
	}
	regParam := registeredParam{
		name:     p.Name,
		usage:    p.Usage,
		newValue: valGetter,
		required: p.Required,
	}
	f.registerParam(regParam)
	return RegisteredStringParam{regParam}
}

// ParamNamePattern describes the regular expression a parameter name must match.
const ParamNamePattern = "^[a-zA-Z0-9][a-zA-Z0-9_-]*$"

var paramNameRegex = regexp.MustCompile(ParamNamePattern)

func (f *Flow) registerParam(p registeredParam) {
	if !paramNameRegex.MatchString(p.name) {
		panic("parameter name must match ParamNamePattern")
	}
	if p.newValue == nil {
		panic("parameter is missing default value factory")
	}
	if _, exists := f.params[p.name]; exists {
		panic(fmt.Sprintf("%s parameter was already registered", p.name))
	}
	if f.params == nil {
		f.params = make(map[string]registeredParam)
	}
	f.params[p.name] = p
}

// TaskNamePattern describes the regular expression a task name must match.
const TaskNamePattern = "^[a-zA-Z0-9_][a-zA-Z0-9_-]*$"

var taskNameRegex = regexp.MustCompile(TaskNamePattern)

// Register registers the task. It panics in case of any error.
func (f *Flow) Register(task Task) RegisteredTask {
	// validate
	if !taskNameRegex.MatchString(task.Name) {
		panic("task name must match TaskNamePattern")
	}
	if f.isRegistered(task.Name) {
		panic(fmt.Sprintf("%s task was already registered", task.Name))
	}
	for _, dep := range task.Deps {
		if !f.isRegistered(dep.task.Name) {
			panic(fmt.Sprintf("invalid dependency %s", dep.task.Name))
		}
	}

	f.tasks[task.Name] = task
	return RegisteredTask{task: task}
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *Flow) Run(ctx context.Context, args ...string) int {
	if ctx == nil {
		ctx = context.Background()
	}

	flow := &flowRunner{
		output:      f.Output,
		params:      f.params,
		tasks:       f.tasks,
		verbose:     f.VerboseParam(),
		workDir:     f.WorkDirParam(),
		defaultTask: f.DefaultTask,
	}

	if flow.output == nil {
		flow.output = os.Stdout
	}

	return flow.Run(ctx, args)
}

func (f *Flow) isRegistered(name string) bool {
	if f.tasks == nil {
		f.tasks = map[string]Task{}
	}
	_, ok := f.tasks[name]
	return ok
}
