package taskflow

import (
	"context"
	"errors"
	"flag"
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

	params map[string]parameter
	tasks  map[string]Task
}

type parameter struct {
	info     ParameterInfo
	register func(*flag.FlagSet)
}

// ParameterInfo represents the general information of a parameter for one or more tasks.
type ParameterInfo struct {
	Name  string
	Usage string
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

func (p RegisteredParam) value(tf *TF) Value {
	value, existing := tf.paramValues[p.name]
	if !existing {
		tf.Fatal(&ParamError{Key: p.name, Err: errors.New("parameter not registered")})
	}
	return value
}

// Value represents an instance of a generic parameter.
// It deliberately matches the signature of type flag.Value as it is used for the
// underlying implementation.
type Value interface {
	String() string
	Set(string) error
}

// ValueParam represents a registered parameter based on a generic implementation.
type ValueParam struct {
	RegisteredParam
}

// Get returns the concrete instance of the generic parameter in the given flow.
func (p ValueParam) Get(tf *TF) Value {
	return p.value(tf)
}

// BoolParam represents a registered boolean parameter.
type BoolParam struct {
	RegisteredParam
}

// Get returns the boolean value of the parameter in the given flow.
func (p BoolParam) Get(tf *TF) bool {
	value := p.value(tf)
	return value.(flag.Getter).Get().(bool)
}

// IntParam represents a registered integer parameter.
type IntParam struct {
	RegisteredParam
}

// Get returns the integer value of the parameter in the given flow.
func (p IntParam) Get(tf *TF) int {
	value := p.value(tf)
	return value.(flag.Getter).Get().(int)
}

// StringParam represents a registered string parameter.
type StringParam struct {
	RegisteredParam
}

// Get returns the string value of the parameter in the given flow.
func (p StringParam) Get(tf *TF) string {
	value := p.value(tf)
	return value.(flag.Getter).Get().(string)
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
	f.configure(func(set *flag.FlagSet) {
		set.Var(newValue(), info.Name, info.Usage)
	}, info)
	return ValueParam{RegisteredParam{name: info.Name}}
}

// ConfigureBool registers a boolean parameter.
func (f *Taskflow) ConfigureBool(defaultValue bool, info ParameterInfo) BoolParam {
	f.configure(func(set *flag.FlagSet) {
		set.Bool(info.Name, defaultValue, info.Usage)
	}, info)
	return BoolParam{RegisteredParam{name: info.Name}}
}

// ConfigureInt registers an integer parameter.
func (f *Taskflow) ConfigureInt(defaultValue int, info ParameterInfo) IntParam {
	f.configure(func(set *flag.FlagSet) {
		set.Int(info.Name, defaultValue, info.Usage)
	}, info)
	return IntParam{RegisteredParam{name: info.Name}}
}

// ConfigureString registers a string parameter.
func (f *Taskflow) ConfigureString(defaultValue string, info ParameterInfo) StringParam {
	f.configure(func(set *flag.FlagSet) {
		set.String(info.Name, defaultValue, info.Usage)
	}, info)
	return StringParam{RegisteredParam{name: info.Name}}
}

func (f *Taskflow) configure(register func(*flag.FlagSet), info ParameterInfo) {
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
		register: register,
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
