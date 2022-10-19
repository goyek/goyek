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
	Output io.Writer // output where text is printed; os.Stdout by default

	// Usage is the function called when an error occurs while parsing tasks.
	// The field is a function that may be changed to point to
	// a custom error handler. By default it calls Print.
	Usage func()

	// Logger used by TF's logging functions. CodeLineLogger by default.
	//
	// TODO: If Helper() is implemented then it is called when TF.Helper() is called.
	Logger Logger

	tasks       map[string]taskSnapshot // snapshot of defined tasks
	defaultTask string                  // task to run when none is explicitly provided
	middlewares []Middleware
}

// Middleware represents a task runner interceptor.
type Middleware func(Runner) Runner

// taskSnapshot is a copy of the task to make the flow usage safer.
type taskSnapshot struct {
	name   string
	usage  string
	deps   []string
	action func(tf *TF)
}

// Tasks returns all tasks sorted in lexicographical order.
func (f *Flow) Tasks() []DefinedTask {
	var tasks []DefinedTask
	for _, task := range f.tasks {
		tasks = append(tasks, registeredTask{task})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name() < tasks[j].Name() })
	return tasks
}

// Define registers the task. It panics in case of any error.
func (f *Flow) Define(task Task) DefinedTask {
	// validate
	if task.Name == "" {
		panic("task name cannot be empty")
	}
	if f.isDefined(task.Name) {
		panic(fmt.Sprintf("task was already defined: %s", task.Name))
	}
	for _, dep := range task.Deps {
		if !f.isDefined(dep.Name()) {
			panic(fmt.Sprintf("dependency was not defined: %s", dep.Name()))
		}
	}

	var deps []string
	for _, dep := range task.Deps {
		deps = append(deps, dep.Name())
	}
	taskCopy := taskSnapshot{
		name:   task.Name,
		usage:  task.Usage,
		deps:   deps,
		action: task.Action,
	}
	f.tasks[task.Name] = taskCopy
	return registeredTask{taskCopy}
}

func (f *Flow) isDefined(name string) bool {
	if f.tasks == nil {
		f.tasks = map[string]taskSnapshot{}
	}
	_, ok := f.tasks[name]
	return ok
}

// Default returns the default task.
// Returns nil of there is no default task.
func (f *Flow) Default() DefinedTask {
	if f.defaultTask == "" {
		return nil
	}
	return registeredTask{f.tasks[f.defaultTask]}
}

// SetDefault sets a task to run when none is explicitly provided.
// It panics in case of any error.
func (f *Flow) SetDefault(task DefinedTask) {
	if !f.isDefined(task.Name()) {
		panic(fmt.Sprintf("task was not defined: %s", task.Name()))
	}
	f.defaultTask = task.Name()
}

// Use adds task runner middlewares (iterceptors).
func (f *Flow) Use(middlewares ...Middleware) {
	for _, m := range middlewares {
		if m == nil {
			panic("middleware cannot be nil")
		}
		f.middlewares = append(f.middlewares, m)
	}
}

// FailError is returned by Flow.Execute when a task failed.
type FailError struct {
	Task string
}

func (err *FailError) Error() string {
	return "task failed: " + err.Task
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
// Returns nil if no task has failed.
// Returns FailError if a task failed.
// Returns other error in case of invalid input or context error.
func (f *Flow) Execute(ctx context.Context, args ...string) error {
	out := f.Output
	if out == nil {
		out = os.Stdout
	}

	logger := f.Logger
	if logger == nil {
		logger = &CodeLineLogger{}
	}

	var tasks []string
	for _, arg := range args {
		if arg == "" {
			return errors.New("task name cannot be empty")
		}
		if _, ok := f.tasks[arg]; !ok {
			return errors.New("task provided but not defined: " + arg)
		}
		tasks = append(tasks, arg)
	}
	if len(tasks) == 0 && f.defaultTask != "" {
		tasks = append(tasks, f.defaultTask)
	}
	if len(tasks) == 0 {
		return errors.New("no task provided")
	}

	var middlewares []Middleware
	middlewares = append(middlewares, f.middlewares...)

	r := &executor{
		output:      out,
		defined:     f.tasks,
		logger:      logger,
		middlewares: middlewares,
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return r.Execute(ctx, tasks)
}

const (
	exitCodePass    = 0
	exitCodeFail    = 1
	exitCodeInvalid = 2
)

// Main runs provided tasks and all their dependencies.
// Each task is executed at most once.
// It exits the current program when after the run is finished
// or SIGINT was send to interrupt the execution.
// 0 exit code means that non of the tasks failed.
// 1 exit code means that a task has failed or the execution was interrupted.
// 2 exit code means that the input was invalid.
// Calls Usage when invalid args are provided.
func (f *Flow) Main(args []string) {
	out := f.Output
	if out == nil {
		out = os.Stdout
	}

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

	// change working directory to repo root (per convention)
	if err := os.Chdir(".."); err != nil {
		fmt.Println(err)
		fmt.Fprintln(out, err)
		os.Exit(exitCodeInvalid)
	}

	exitCode := f.main(ctx, args...)
	os.Exit(exitCode)
}

func (f *Flow) main(ctx context.Context, args ...string) int {
	out := f.Output
	if out == nil {
		out = os.Stdout
	}

	from := time.Now()
	err := f.Execute(ctx, args...)
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
		if f.Usage != nil {
			f.Usage()
		} else {
			f.Print()
		}
		return exitCodeInvalid
	}
	fmt.Fprintf(out, "ok\t%.3fs\n", time.Since(from).Seconds())
	return exitCodePass
}

// Print prints, to os.Stdout unless configured otherwise,
// the information about the registered tasks.
// Tasks with empty Usage are not printed.
func (f *Flow) Print() {
	out := f.Output
	if out == nil {
		out = os.Stdout
	}

	if f.defaultTask != "" {
		fmt.Fprintf(out, "Default task: %s\n", f.defaultTask)
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
			deps = " (depends on: " + strings.Join(task.Deps(), ", ") + ")"
		}
		fmt.Fprintf(w, "  %s\t%s\n", task.Name(), task.Usage()+deps)
	}
	w.Flush() //nolint:errcheck,gosec // not checking errors when writing to output
}
