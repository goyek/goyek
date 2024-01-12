package goyek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// Flow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
type Flow struct {
	output io.Writer
	usage  func()
	logger Logger // TODO: If Helper() is implemented then it is called when A.Helper() is called.

	tasks       map[string]*taskSnapshot // snapshot of defined tasks
	defaultTask *taskSnapshot            // task to run when none is explicitly provided
	middlewares []Middleware
}

// DefaultFlow is the default flow.
// The top-level functions such as Define, Main, and so on are wrappers for the methods of Flow.
var DefaultFlow = &Flow{}

// Middleware represents a task runner interceptor.
type Middleware func(Runner) Runner

// taskSnapshot is a copy of the task to make the flow usage safer.
type taskSnapshot struct {
	name     string
	usage    string
	deps     []*taskSnapshot
	action   func(a *A)
	parallel bool
}

// Tasks returns all tasks sorted in lexicographical order.
func Tasks() []*DefinedTask {
	return DefaultFlow.Tasks()
}

// Tasks returns all tasks sorted in lexicographical order.
func (f *Flow) Tasks() []*DefinedTask {
	var tasks []*DefinedTask
	for _, task := range f.tasks {
		tasks = append(tasks, &DefinedTask{task, f})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name() < tasks[j].Name() })
	return tasks
}

// Define registers the task. It panics in case of any error.
func Define(task Task) *DefinedTask {
	return DefaultFlow.Define(task)
}

// Define registers the task. It panics in case of any error.
func (f *Flow) Define(task Task) *DefinedTask {
	// validate
	if task.Name == "" {
		panic("task name cannot be empty")
	}
	if f.isDefined(task.Name, f) {
		panic("task with the same name is already defined")
	}
	for _, dep := range task.Deps {
		if !f.isDefined(dep.Name(), dep.flow) {
			panic("dependency was not defined: " + dep.Name())
		}
	}

	var deps []*taskSnapshot
	for _, dep := range task.Deps {
		deps = append(deps, dep.taskSnapshot)
	}
	taskCopy := &taskSnapshot{
		name:     task.Name,
		usage:    task.Usage,
		deps:     deps,
		action:   task.Action,
		parallel: task.Parallel,
	}
	f.tasks[task.Name] = taskCopy
	return &DefinedTask{taskCopy, f}
}

// Undefine unregisters the task. It panics in case of any error.
func Undefine(task *DefinedTask) {
	DefaultFlow.Undefine(task)
}

// Undefine unregisters the task. It panics in case of any error.
func (f *Flow) Undefine(task *DefinedTask) {
	snapshot := task.taskSnapshot
	if !f.isDefined(snapshot.name, task.flow) {
		panic("task was not defined: " + snapshot.name)
	}

	delete(f.tasks, snapshot.name)

	for _, task := range f.tasks {
		if len(task.deps) == 0 {
			continue
		}
		var cleanDep []*taskSnapshot
		for _, dep := range task.deps {
			if dep == snapshot {
				continue
			}
			cleanDep = append(cleanDep, dep)
		}
		task.deps = cleanDep
	}

	if f.defaultTask == snapshot {
		f.defaultTask = nil
	}
}

func (f *Flow) isDefined(name string, flow *Flow) bool {
	if f.tasks == nil {
		f.tasks = map[string]*taskSnapshot{}
	}
	if f != flow {
		return false // defined in other flow
	}
	_, ok := f.tasks[name]
	return ok
}

// Output returns the destination used for printing messages.
// [os.Stdout] is returned if output was not set or was set to nil.
func Output() io.Writer {
	return DefaultFlow.Output()
}

// Output returns the destination used for printing messages.
// [os.Stdout] is returned if output was not set or was set to nil.
func (f *Flow) Output() io.Writer {
	if f.output == nil {
		return os.Stdout
	}
	return f.output
}

// SetOutput sets the output destination.
func SetOutput(out io.Writer) {
	DefaultFlow.SetOutput(out)
}

// SetOutput sets the output destination.
func (f *Flow) SetOutput(out io.Writer) {
	f.output = out
}

// GetLogger returns the logger used by A's logging functions
// [CodeLineLogger] is returned if logger was not set or was set to nil.
func GetLogger() Logger {
	return DefaultFlow.Logger()
}

// Logger returns the logger used by A's logging functions
// [CodeLineLogger] is returned if logger was not set or was set to nil.
func (f *Flow) Logger() Logger {
	if f.logger == nil {
		return &CodeLineLogger{}
	}
	return f.logger
}

// SetLogger sets the logger used by A's logging functions.
//
// [A] uses following methods if implemented:
//
//	Error(w io.Writer, args ...interface{})
//	Errorf(w io.Writer, format string, args ...interface{})
//	Fatal(w io.Writer, args ...interface{})
//	Fatalf(w io.Writer, format string, args ...interface{})
//	Skip(w io.Writer, args ...interface{})
//	Skipf(w io.Writer, format string, args ...interface{})
//	Helper()
func SetLogger(logger Logger) {
	DefaultFlow.SetLogger(logger)
}

// SetLogger sets the logger used by A's logging functions.
//
// [A] uses following methods if implemented:
//
//	Error(w io.Writer, args ...interface{})
//	Errorf(w io.Writer, format string, args ...interface{})
//	Fatal(w io.Writer, args ...interface{})
//	Fatalf(w io.Writer, format string, args ...interface{})
//	Skip(w io.Writer, args ...interface{})
//	Skipf(w io.Writer, format string, args ...interface{})
//	Helper()
func (f *Flow) SetLogger(logger Logger) {
	f.logger = logger
}

// Usage returns a function that prints a usage message documenting the flow.
// It is called when an error occurs while parsing the flow.
// [Print] is returned if a function was not set or was set to nil.
func Usage() func() {
	return DefaultFlow.Usage()
}

// Usage returns a function that prints a usage message documenting the flow.
// It is called when an error occurs while parsing the flow.
// [Flow.Print] is returned if a function was not set or was set to nil.
func (f *Flow) Usage() func() {
	if f.usage == nil {
		return f.Print
	}
	return f.usage
}

// SetUsage sets the function called when an error occurs while parsing tasks.
func SetUsage(fn func()) {
	DefaultFlow.SetUsage(fn)
}

// SetUsage sets the function called when an error occurs while parsing tasks.
func (f *Flow) SetUsage(fn func()) {
	f.usage = fn
}

// Default returns the default task.
// nil is returned if default was not set.
func Default() *DefinedTask {
	return DefaultFlow.Default()
}

// Default returns the default task.
// nil is returned if default was not set.
func (f *Flow) Default() *DefinedTask {
	if f.defaultTask == nil {
		return nil
	}
	return &DefinedTask{f.defaultTask, f}
}

// SetDefault sets a task to run when none is explicitly provided.
// It panics in case of any error.
func SetDefault(task *DefinedTask) {
	DefaultFlow.SetDefault(task)
}

// SetDefault sets a task to run when none is explicitly provided.
// Passing nil clears the default task.
// It panics in case of any error.
func (f *Flow) SetDefault(task *DefinedTask) {
	if task == nil {
		f.defaultTask = nil
		return
	}

	if !f.isDefined(task.Name(), task.flow) {
		panic("task was not defined: " + task.Name())
	}
	f.defaultTask = task.taskSnapshot
}

// Use adds task runner middlewares (interceptors).
func Use(middlewares ...Middleware) {
	DefaultFlow.Use(middlewares...)
}

// Use adds task runner middlewares (interceptors).
func (f *Flow) Use(middlewares ...Middleware) {
	for _, m := range middlewares {
		if m == nil {
			panic("middleware cannot be nil")
		}
		f.middlewares = append(f.middlewares, m)
	}
}

// Option configures the flow execution.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	noDeps    bool
	skipTasks []string
}

// NoDeps is an option to skip processing of all dependencies.
func NoDeps() Option {
	return optionFunc(func(c *config) {
		c.noDeps = true
	})
}

// Skip is an option to skip processing of given tasks.
func Skip(tasks ...string) Option {
	return optionFunc(func(c *config) {
		c.skipTasks = append(c.skipTasks, tasks...)
	})
}

// FailError pointer is returned by [Flow.Execute] when a task failed.
type FailError struct {
	Task string
}

func (err *FailError) Error() string {
	return "task failed: " + err.Task
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
// Returns nil if no task has failed,
// [FailError] if a task failed,
// other errors in case of invalid input or context error.
func Execute(ctx context.Context, tasks []string, opts ...Option) error {
	return DefaultFlow.Execute(ctx, tasks, opts...)
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
// Returns nil if no task has failed,
// [FailError] if a task failed,
// other errors in case of invalid input or context error.
func (f *Flow) Execute(ctx context.Context, tasks []string, opts ...Option) error {
	for _, task := range tasks {
		if task == "" {
			return errors.New("task name cannot be empty")
		}
		if _, ok := f.tasks[task]; !ok {
			return errors.New("task provided but not defined: " + task)
		}
	}
	if len(tasks) == 0 && f.defaultTask != nil {
		tasks = append(tasks, f.defaultTask.name)
	}
	if len(tasks) == 0 {
		return errors.New("no task provided")
	}

	var middlewares []Middleware
	middlewares = append(middlewares, f.middlewares...)

	cfg := &config{}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	for _, skippedTask := range cfg.skipTasks {
		if skippedTask == "" {
			return errors.New("skipped task name cannot be empty")
		}
		if _, ok := f.tasks[skippedTask]; !ok {
			return errors.New("skipped task provided but not defined: " + skippedTask)
		}
	}

	r := &executor{
		output:      f.Output(),
		defined:     f.tasks,
		logger:      f.Logger(),
		middlewares: middlewares,
		noDeps:      cfg.noDeps,
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return r.Execute(ctx, tasks, cfg.skipTasks)
}

const (
	exitCodePass    = 0
	exitCodeFail    = 1
	exitCodeInvalid = 2
)

// Main runs provided tasks and all their dependencies.
// Each task is executed at most once.
// It exits the current program when after the run is finished
// or SIGINT interrupted the execution.
//   - 0 exit code means that non of the tasks failed.
//   - 1 exit code means that a task has failed or the execution was interrupted.
//   - 2 exit code means that the input was invalid.
//
// Calls [Usage] when invalid args are provided.
func Main(args []string, opts ...Option) {
	DefaultFlow.Main(args, opts...)
}

// Main runs provided tasks and all their dependencies.
// Each task is executed at most once.
// It exits the current program when after the run is finished
// or SIGINT interrupted the execution.
//   - 0 exit code means that non of the tasks failed.
//   - 1 exit code means that a task has failed or the execution was interrupted.
//   - 2 exit code means that the input was invalid.
//
// Calls [Usage] when invalid args are provided.
func (f *Flow) Main(args []string, opts ...Option) {
	out := f.Output()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // first signal, cancel context
		fmt.Fprintln(out, "first interrupt, graceful stop")
		cancel()

		<-c // second signal, hard exit
		fmt.Fprintln(out, "second interrupt, exit")
		os.Exit(exitCodeFail)
	}()

	exitCode := f.main(ctx, args, opts...)
	os.Exit(exitCode)
}

func (f *Flow) main(ctx context.Context, args []string, opts ...Option) int {
	out := f.Output()

	from := time.Now()
	err := f.Execute(ctx, args, opts...)
	if _, ok := err.(*FailError); ok {
		fmt.Fprintf(out, "%v\t%.3fs\n", err, time.Since(from).Seconds())
		return exitCodeFail
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		fmt.Fprintf(out, "%v\t%.3fs\n", err, time.Since(from).Seconds())
		return exitCodeFail
	}
	if err != nil {
		fmt.Fprintln(out, err.Error())
		f.Usage()()
		return exitCodeInvalid
	}
	fmt.Fprintf(out, "ok\t%.3fs\n", time.Since(from).Seconds())
	return exitCodePass
}

// Print prints the information about the registered tasks.
// Tasks with empty [Task.Usage] are not printed.
func Print() {
	DefaultFlow.Print()
}

// Print prints the information about the registered tasks.
// Tasks with empty [Task.Usage] are not printed.
func (f *Flow) Print() {
	out := f.Output()

	if f.defaultTask != nil {
		fmt.Fprintf(out, "Default task: %s\n", f.defaultTask.name)
	}

	fmt.Fprintln(out, "Tasks:")
	var (
		minwidth      = 5
		tabwidth      = 0
		padding       = 2
		padchar  byte = ' '
	)
	w := tabwriter.NewWriter(out, minwidth, tabwidth, padding, padchar, 0)
	for _, task := range f.Tasks() {
		if task.Usage() == "" {
			continue
		}
		deps := ""
		if len(task.Deps()) > 0 {
			depNames := make([]string, 0, len(task.Deps()))
			for _, dep := range task.Deps() {
				depNames = append(depNames, dep.Name())
			}
			deps = " (depends on: " + strings.Join(depNames, ", ") + ")"
		}
		fmt.Fprintf(w, "  %s\t%s\n", task.Name(), task.Usage()+deps)
	}
	w.Flush() //nolint:errcheck // not checking errors when writing to output
}
