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
	"sync"
	"text/tabwriter"
)

// Flow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
type Flow struct {
	mu     sync.RWMutex
	output io.Writer
	usage  func()
	logger Logger

	tasks               map[string]*DefinedTask // snapshot of defined tasks
	defaultTask         *DefinedTask            // task to run when none is explicitly provided
	middlewares         []Middleware
	executorMiddlewares []ExecutorMiddleware
}

// DefaultFlow is the default flow.
// The top-level functions such as Define, Main, and so on are wrappers for the methods of Flow.
var DefaultFlow = &Flow{}

// Tasks returns all tasks sorted in lexicographical order.
func Tasks() []*DefinedTask {
	return DefaultFlow.Tasks()
}

// Tasks returns all tasks sorted in lexicographical order.
func (f *Flow) Tasks() []*DefinedTask {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var tasks []*DefinedTask
	for _, task := range f.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].name < tasks[j].name })
	return tasks
}

// Define registers the task. It panics in case of any error.
func Define(task Task) *DefinedTask {
	return DefaultFlow.Define(task)
}

// Define registers the task. It panics in case of any error.
func (f *Flow) Define(task Task) *DefinedTask {
	f.mu.Lock()
	defer f.mu.Unlock()
	// validate
	if task.Name == "" {
		panic("task name cannot be empty")
	}
	if f.isDefinedLocked(task.Name, f) {
		panic("task with the same name is already defined")
	}
	for _, dep := range task.Deps {
		if !f.isDefinedLocked(dep.name, dep.flow) {
			panic("dependency was not defined: " + dep.name)
		}
	}

	taskCopy := &DefinedTask{
		name:     task.Name,
		usage:    task.Usage,
		deps:     task.Deps,
		action:   task.Action,
		parallel: task.Parallel,
		flow:     f,
	}
	f.tasks[task.Name] = taskCopy
	return taskCopy
}

// Undefine unregisters the task. It panics in case of any error.
func Undefine(task *DefinedTask) {
	DefaultFlow.Undefine(task)
}

// Undefine unregisters the task. It panics in case of any error.
func (f *Flow) Undefine(task *DefinedTask) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.isDefinedLocked(task.name, task.flow) {
		panic("task was not defined: " + task.name)
	}

	delete(f.tasks, task.name)

	for _, t := range f.tasks {
		if len(t.deps) == 0 {
			continue
		}
		var cleanDep []*DefinedTask
		for _, dep := range t.deps {
			if dep == task {
				continue
			}
			cleanDep = append(cleanDep, dep)
		}
		t.deps = cleanDep
	}

	if f.defaultTask == task {
		f.defaultTask = nil
	}
}

func (f *Flow) isDefined(name string, flow *Flow) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.isDefinedLocked(name, flow)
}

func (f *Flow) isDefinedLocked(name string, flow *Flow) bool {
	if f.tasks == nil {
		f.tasks = map[string]*DefinedTask{}
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
	f.mu.RLock()
	defer f.mu.RUnlock()
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
	f.mu.Lock()
	defer f.mu.Unlock()
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
	f.mu.RLock()
	defer f.mu.RUnlock()
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
	f.mu.Lock()
	defer f.mu.Unlock()
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
	f.mu.RLock()
	defer f.mu.RUnlock()
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
	f.mu.Lock()
	defer f.mu.Unlock()
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
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.defaultTask
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
	f.mu.Lock()
	defer f.mu.Unlock()
	if task == nil {
		f.defaultTask = nil
		return
	}

	if !f.isDefinedLocked(task.name, task.flow) {
		panic("task was not defined: " + task.name)
	}
	f.defaultTask = task
}

// Use adds task runner middlewares (interceptors).
func Use(middlewares ...Middleware) {
	DefaultFlow.Use(middlewares...)
}

// Use adds task runner middlewares (interceptors).
func (f *Flow) Use(middlewares ...Middleware) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range middlewares {
		if m == nil {
			panic("middleware cannot be nil")
		}
		f.middlewares = append(f.middlewares, m)
	}
}

// UseExecutor adds flow executor middlewares (interceptors).
func UseExecutor(middlewares ...ExecutorMiddleware) {
	DefaultFlow.UseExecutor(middlewares...)
}

// UseExecutor adds flow executor middlewares (interceptors).
func (f *Flow) UseExecutor(middlewares ...ExecutorMiddleware) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range middlewares {
		if m == nil {
			panic("middleware cannot be nil")
		}
		f.executorMiddlewares = append(f.executorMiddlewares, m)
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
// [*FailError] if a task failed,
// other errors in case of invalid input or context error.
func Execute(ctx context.Context, tasks []string, opts ...Option) error {
	return DefaultFlow.Execute(ctx, tasks, opts...)
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
// Returns nil if no task has failed,
// [*FailError] if a task failed,
// other errors in case of invalid input or context error.
func (f *Flow) Execute(ctx context.Context, tasks []string, opts ...Option) error {
	f.mu.RLock()
	// snapshot
	defined := make(map[string]*DefinedTask, len(f.tasks))
	for k, v := range f.tasks {
		defined[k] = v
	}
	middlewares := make([]Middleware, len(f.middlewares))
	copy(middlewares, f.middlewares)
	executorMiddlewares := make([]ExecutorMiddleware, len(f.executorMiddlewares))
	copy(executorMiddlewares, f.executorMiddlewares)
	defaultTask := f.defaultTask
	output := f.output
	logger := f.logger
	f.mu.RUnlock()

	cfg := &config{}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	// prepare runner
	r := &executor{
		defined:     defined,
		middlewares: middlewares,
		defaultTask: defaultTask,
	}
	runner := r.Execute

	// apply defined executor middlewares
	for _, middleware := range executorMiddlewares {
		runner = middleware(runner)
	}

	if output == nil {
		output = os.Stdout
	}
	if logger == nil {
		logger = &CodeLineLogger{}
	}

	in := ExecuteInput{
		Context:   ctx,
		Tasks:     tasks,
		SkipTasks: cfg.skipTasks,
		NoDeps:    cfg.noDeps,
		Output:    output,
		Logger:    logger,
	}
	return runner(in)
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
	err := f.Execute(ctx, args, opts...)
	var ferr *FailError
	if errors.As(err, &ferr) {
		return exitCodeFail
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return exitCodeFail
	}
	if err != nil {
		f.Usage()()
		return exitCodeInvalid
	}
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
	f.mu.RLock()
	out := f.output
	if out == nil {
		out = os.Stdout
	}
	defaultTask := f.defaultTask
	tasks := make([]*DefinedTask, 0, len(f.tasks))
	for _, task := range f.tasks {
		tasks = append(tasks, task)
	}
	f.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name() < tasks[j].Name() })

	if defaultTask != nil {
		fmt.Fprintf(out, "Default task: %s\n", defaultTask.name)
	}

	fmt.Fprintln(out, "Tasks:")
	var (
		minwidth      = 5
		tabwidth      = 0
		padding       = 2
		padchar  byte = ' '
	)
	w := tabwriter.NewWriter(out, minwidth, tabwidth, padding, padchar, 0)
	for _, task := range tasks {
		usage := task.Usage()
		if usage == "" {
			continue
		}
		deps := ""
		taskDeps := task.Deps()
		if len(taskDeps) > 0 {
			depNames := make([]string, 0, len(taskDeps))
			for _, dep := range taskDeps {
				depNames = append(depNames, dep.Name())
			}
			deps = " (depends on: " + strings.Join(depNames, ", ") + ")"
		}
		fmt.Fprintf(w, "  %s\t%s\n", task.Name(), usage+deps)
	}
	w.Flush()
}
