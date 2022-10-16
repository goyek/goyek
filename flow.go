package goyek

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"text/tabwriter"
)

const (
	// CodePass indicates that no task has failed.
	CodePass = 0
	// CodeFail indicates that a task has failed or the run was interrupted.
	CodeFail = 1
	// CodeInvalidArgs indicates that an error occurerd while parsing tasks.
	CodeInvalidArgs = 2
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

	// Logger used by TF's logging functions to decorate the text.
	// CodeLineLogDecorator by default.
	//
	// TODO: If Helper() is implemented then it is called when TF.Helper() is called.
	Logger Logger

	tasks       map[string]taskSnapshot // snapshot of defined tasks
	defaultTask string                  // task to run when none is explicitly provided
	middlewares []func(Runner) Runner
}

// taskSnapshot is a copy of the task to make the flow usage safer.
type taskSnapshot struct {
	name   string
	usage  string
	deps   []string
	action func(tf *TF)
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

// SetDefault sets a task to run when none is explicitly provided.
// It panics in case of any error.
func (f *Flow) SetDefault(task DefinedTask) {
	if !f.isDefined(task.Name()) {
		panic(fmt.Sprintf("task was not defined: %s", task.Name()))
	}
	f.defaultTask = task.Name()
}

// Use banana.
func (f *Flow) Use(middlewares ...func(Runner) Runner) {
	for _, m := range middlewares {
		if m == nil {
			panic("middleware cannot be nil")
		}
		f.middlewares = append(f.middlewares, m)
	}
}

// Execute runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *Flow) Execute(ctx context.Context, args ...string) int {
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
			fmt.Fprintln(out, "task name cannot be empty") // TODO: move to Main
			return f.invalid()
		}
		if _, ok := f.tasks[arg]; !ok {
			fmt.Fprintf(out, "task provided but not defined: %s\n", arg) // TODO: move to Main
			return f.invalid()
		}
		tasks = append(tasks, arg)
	}
	if len(tasks) == 0 && f.defaultTask != "" {
		tasks = append(tasks, f.defaultTask)
	}
	if len(tasks) == 0 {
		fmt.Fprintln(out, "no task provided") // TODO: move to Main
		return f.invalid()
	}

	var middlewares []func(Runner) Runner
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
	if !r.Execute(ctx, tasks) {
		return CodeFail
	}
	return CodePass
}

func (f *Flow) invalid() int {
	if f.Usage != nil {
		f.Usage()
	} else {
		f.Print()
	}
	return CodeInvalidArgs
}

// Main runs provided tasks and all their dependencies.
// Each task is executed at most once.
// It exits the current program when after the run is finished
// or SIGINT was send to interrupt the execution.
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
		os.Exit(CodeFail)
	}()

	// change working directory to repo root (per convention)
	if err := os.Chdir(".."); err != nil {
		fmt.Println(err)
		fmt.Fprintln(out, err)
		os.Exit(CodeInvalidArgs)
	}

	// run flow
	exitCode := f.Execute(ctx, args...)
	os.Exit(exitCode)
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

// Default returns the default task.
// Returns nil of there is no default task.
func (f *Flow) Default() DefinedTask {
	if f.defaultTask == "" {
		return nil
	}
	return registeredTask{f.tasks[f.defaultTask]}
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
		minwidth      = 3
		tabwidth      = 1
		padding       = 3
		padchar  byte = ' '
	)
	w := tabwriter.NewWriter(out, minwidth, tabwidth, padding, padchar, 0)
	for _, task := range f.Tasks() {
		if task.Usage() == "" {
			continue
		}
		fmt.Fprintf(w, "\t%s\t%s\t%s\n", task.Name(), task.Usage(), strings.Join(task.Deps(), ", "))
	}
	w.Flush() //nolint:errcheck,gosec // not checking errors when writing to output
}
