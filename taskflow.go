/*
Package taskflow helps implementing build automation.
It is intended to be used in concert with the "go run" command,
to run a program which implements the build pipeline (called taskflow).
A taskflow consists of a set of registered tasks.
A task has a name, can have a defined command, which is a function with signature
	func (*taskflow.TF)
and can have dependencies (already defined tasks).

When the taskflow is executed for given tasks,
then the tasks' commands are run in the order defined by the dependencies.
The tasks dependencies are run in a recusrive manner, however each is going to be run at most once.

The taskflow is interupted in case a command fails.
Within these functions, use the Error, Fail or related methods to signal failure.
*/
package taskflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	// ErrTaskNotRegistered is the error returned if the the task
	// that is requested to run is not registered.
	ErrTaskNotRegistered = errors.New("task provided but not registered")

	// ErrTaskFail is the error returned if a task command failed.
	ErrTaskFail = errors.New("task failed")
)

// Taskflow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
// By default Taskflow prints to Stdout, but it can be change by setting Out.
type Taskflow struct {
	Verbose bool
	Out     io.Writer

	tasks map[string]Task
}

// RegisteredTask represents a task that has been registered to a Taskflow.
// It can be used as a dependency for another Task.
type RegisteredTask struct {
	name string
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
func (f *Taskflow) Run(ctx context.Context, taskNames ...string) error {
	// validate
	for _, name := range taskNames {
		if !f.isRegistered(name) {
			return ErrTaskNotRegistered
		}
	}

	// recursive run
	executedTasks := map[string]bool{}
	for _, name := range taskNames {
		if err := f.run(ctx, name, executedTasks); err != nil {
			return err
		}
	}
	return nil
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
		return ErrTaskFail
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

	runner := Runner{
		Ctx:  ctx,
		Name: task.Name,
		Out:  w,
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
	if f.Out == nil {
		return os.Stdout
	}
	return f.Out
}

func reportTaskStart(taskName string) string {
	return fmt.Sprintf("===== TASK  %s\n", taskName)
}

func reportTaskEnd(status string, taskName string, d time.Duration) string {
	return fmt.Sprintf("----- %s: %s (%.2fs)\n", status, taskName, d.Seconds())
}
