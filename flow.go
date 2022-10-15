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
	Output  io.Writer // output where text is printed; os.Stdout by default
	Verbose bool      // control the printing

	// Usage is the function called when an error occurs while parsing tasks.
	// The field is a function that may be changed to point to
	// a custom error handler. By default it calls Print.
	Usage func()

	tasks       map[string]taskSnapshot // snapshot of defined tasks
	defaultTask string                  // task to run when none is explicitly provided
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

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *Flow) Run(ctx context.Context, args ...string) int {
	out := f.Output
	if out == nil {
		out = os.Stdout
	}

	var tasks []string
	for _, arg := range args {
		if arg == "" {
			fmt.Fprintln(out, "task name cannot be empty")
			return f.invalid()
		}
		if _, ok := f.tasks[arg]; !ok {
			fmt.Fprintf(out, "task provided but not defined: %s\n", arg)
			return f.invalid()
		}
		tasks = append(tasks, arg)
	}
	if len(tasks) == 0 && f.defaultTask != "" {
		tasks = append(tasks, f.defaultTask)
	}
	if len(tasks) == 0 {
		fmt.Fprintln(out, "no task provided")
		return f.invalid()
	}

	r := &flowRunner{
		output:  out,
		defined: f.tasks,
		verbose: f.Verbose,
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if !r.Run(ctx, tasks) {
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
	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // first signal, cancel context
		fmt.Fprintln(f.Output, "first interrupt, graceful stop")
		cancel()

		<-c // second signal, hard exit
		fmt.Fprintln(f.Output, "second interrupt, exit")
		os.Exit(CodeFail)
	}()

	// run flow
	exitCode := f.Run(ctx, args...)
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
		fmt.Fprintf(w, "\t%s\t%s\t%s\n", task.Name(), task.Usage(), strings.Join(task.Deps(), ", "))
	}
	w.Flush() //nolint:errcheck,gosec // not checking errors when writing to output
}
