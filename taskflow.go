/*
Package taskflow helps implementing build automation.
It is intended to be used in concert with the "go run" command,
to run a program which implements the build pipeline (called taskflow).
A taskflow consists of a set of registered tasks.
A task has a name, can have a defined command, which is a function with signature
	func (*taskflow.TF)
and can have dependencies (already defined tasks).

When the taskflow is executed for given tasks,
then the tasks' commands are run in the order defined by their dependencies.
The task's dependencies are run in a recusrive manner, however each is going to be run at most once.

The taskflow is interupted in case a command fails.
Within these functions, use the Error, Fail or related methods to signal failure.
*/
package taskflow

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	// CodePass indicates that taskflow passed.
	CodePass = 0
	// CodeFailure indicates that taskflow failed.
	CodeFailure = 1
	// CodeInvalidArgs indicates that got invalid input.
	CodeInvalidArgs = 2
)

// Taskflow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
// By default Taskflow prints to Stdout, but it can be change by setting Output.
type Taskflow struct {
	Verbose bool
	Output  io.Writer
	Params  Params

	tasks map[string]Task
}

// RegisteredTask represents a task that has been registered to a Taskflow.
// It can be used as a dependency for another Task.
type RegisteredTask struct {
	name string
}

// Params represents Taskflow parameters used within Taskflow.
// The default values set in the struct are overridden in Run method.
type Params map[string]string

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
	// parse args
	cli := flag.NewFlagSet("", flag.ContinueOnError)
	cli.SetOutput(f.Output)
	verbose := cli.Bool("v", false, "Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.")
	usage := func() {
		fmt.Fprintf(cli.Output(), "Usage: [flag(s)] task(s)\n")
		fmt.Fprintf(cli.Output(), "Flags:\n")
		cli.PrintDefaults()

		fmt.Fprintf(cli.Output(), "Tasks:\n")
		w := tabwriter.NewWriter(cli.Output(), 1, 1, 4, ' ', 0)
		keys := make([]string, 0, len(f.tasks))
		for k, task := range f.tasks {
			if task.Description == "" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			t := f.tasks[k]
			fmt.Fprintf(w, "  %s\t%s\n", t.Name, t.Description)
		}
		if err := w.Flush(); err != nil {
			panic(err)
		}
	}
	cli.Usage = usage
	if err := cli.Parse(args); err != nil {
		fmt.Fprintln(cli.Output(), err)
		return CodeInvalidArgs
	}
	if *verbose {
		f.Verbose = true
	}
	if f.Params == nil {
		f.Params = Params{}
	}

	// parse non-flag args
	var tasks []string
	for _, arg := range cli.Args() {
		if paramAssignmentIdx := strings.IndexRune(arg, '='); paramAssignmentIdx > 0 {
			// parameter assignement via 'key=val'
			key := arg[0:paramAssignmentIdx]
			val := arg[paramAssignmentIdx+1:]
			f.Params[key] = val
			continue
		}
		if !f.isRegistered(arg) {
			// task is not registered
			fmt.Fprintf(cli.Output(), "%s task is not registered\n", arg)
			usage()
			return CodeInvalidArgs
		}
		tasks = append(tasks, arg)
	}

	if len(tasks) == 0 {
		fmt.Fprintln(cli.Output(), "no task provided")
		usage()
		return CodeInvalidArgs
	}

	// recursive run
	from := time.Now()
	executedTasks := map[string]bool{}
	for _, name := range tasks {
		if err := f.run(ctx, name, executedTasks); err != nil {
			fmt.Fprintf(cli.Output(), "%v\t%.3fs\n", err, time.Since(from).Seconds())
			return CodeFailure
		}
	}
	fmt.Fprintf(cli.Output(), "ok\t%.3fs\n", time.Since(from).Seconds())
	return CodePass
}

func (f *Taskflow) run(ctx context.Context, name string, executed map[string]bool) error {
	task := f.tasks[name]
	if executed[name] {
		return nil
	}
	for _, dep := range task.Dependencies {
		if err := f.run(ctx, dep.name, executed); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	passed := f.runTask(ctx, task)
	if err := ctx.Err(); err != nil {
		return err
	}
	if !passed {
		return errors.New("task failed")
	}
	executed[name] = true
	return nil
}

func (f *Taskflow) runTask(ctx context.Context, task Task) bool {
	if task.Command == nil {
		return true
	}

	w := f.output()
	if !f.Verbose {
		w = &strings.Builder{}
	}

	_, err := io.WriteString(w, reportTaskStart(task.Name))
	if err != nil {
		panic(err)
	}

	// copy Params to make sure that it cannot be modified by any task's command
	params := make(Params, len(f.Params))
	for k, v := range f.Params {
		params[k] = v
	}

	runner := Runner{
		Ctx:     ctx,
		Name:    task.Name,
		Verbose: f.Verbose,
		Params:  params,
		Output:  w,
	}
	result := runner.Run(task.Command)

	switch {
	default:
		_, err = io.WriteString(w, reportTaskEnd("PASS", task.Name, result.Duration()))
	case result.Failed():
		_, err = io.WriteString(w, reportTaskEnd("FAIL", task.Name, result.Duration()))
	case result.Skipped():
		_, err = io.WriteString(w, reportTaskEnd("SKIP", task.Name, result.Duration()))
	}
	if err != nil {
		panic(err)
	}

	if sb, ok := w.(*strings.Builder); ok && result.failed {
		if _, err := io.Copy(f.output(), strings.NewReader(sb.String())); err != nil {
			panic(err)
		}
	}

	return !result.failed
}

func (f *Taskflow) isRegistered(name string) bool {
	if f.tasks == nil {
		f.tasks = map[string]Task{}
	}
	_, ok := f.tasks[name]
	return ok
}

func (f *Taskflow) output() io.Writer {
	if f.Output == nil {
		return os.Stdout
	}
	return f.Output
}

func reportTaskStart(taskName string) string {
	return fmt.Sprintf("===== TASK  %s\n", taskName)
}

func reportTaskEnd(status string, taskName string, d time.Duration) string {
	return fmt.Sprintf("----- %s: %s (%.2fs)\n", status, taskName, d.Seconds())
}
