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
	// CodeFail indicates that some task has failed or the run was interrupted.
	CodeFail = 1
	// CodeInvalidArgs indicates that an error occurerd while parsing tasks.
	CodeInvalidArgs = 2
)

// Flow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
type Flow struct {
	Output      io.Writer   // output where text is printed; os.Stdout by default
	Verbose     bool        // control the printing
	DefaultTask DefinedTask // task to run when none is explicitly provided

	// Usage is the function called when an error occurs while parsing tasks.
	// The field is a function that may be changed to point to
	// a custom error handler. By default it calls Print.
	Usage func()

	tasks map[string]taskSnapshot
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
	if f.isRegistered(task.Name) {
		panic(fmt.Sprintf("task was already defined: %s", task.Name))
	}
	for _, dep := range task.Deps {
		if !f.isRegistered(dep.Name()) {
			panic(fmt.Sprintf("invalid dependency: %s", dep.Name()))
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

func (f *Flow) isRegistered(name string) bool {
	if f.tasks == nil {
		f.tasks = map[string]taskSnapshot{}
	}
	_, ok := f.tasks[name]
	return ok
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *Flow) Run(ctx context.Context, args ...string) int {
	if ctx == nil {
		ctx = context.Background()
	}

	r := &runner{
		output:  f.Output,
		tasks:   f.tasks,
		verbose: f.Verbose,
	}

	if f.DefaultTask != nil {
		r.defaultTask = f.DefaultTask.Name()
	}

	if r.output == nil {
		r.output = os.Stdout
	}

	exitCode := r.Run(ctx, args)
	if exitCode == CodeInvalidArgs {
		if f.Usage != nil {
			f.Usage()
		} else {
			f.Print()
		}
	}
	return exitCode
}

// Main parses the args and runs the provided tasks.
// It exists when after the flow finished or SIGINT
// was send to interrupt the execution.
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

// Print prints, to os.Stdout unless configured otherwise,
// the information about the registered tasks.
func (f *Flow) Print() {
	out := f.Output
	if out == nil {
		out = os.Stdout
	}

	if f.DefaultTask != nil {
		fmt.Fprintf(out, "Default task: %s\n", f.DefaultTask.Name())
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
	w.Flush() //nolint // not checking errors when writing to output
}
