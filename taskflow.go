package taskflow

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Taskflow TODO.
type Taskflow struct {
	Verbose bool
	Output  io.Writer

	tasks map[string]Task
}

// Dependency TODO.
type Dependency struct {
	name string
}

// Register TODO.
func (f *Taskflow) Register(task Task) (Dependency, error) {
	if f.isRegistered(task.Name) {
		return Dependency{}, fmt.Errorf("%s task was already registered", task.Name) //nolint:goerr113 // TODO
	}
	f.tasks[task.Name] = task
	return Dependency{name: task.Name}, nil
}

// MustRegister TODO.
func (f *Taskflow) MustRegister(task Task) Dependency {
	dep, err := f.Register(task)
	if err != nil {
		panic(err)
	}
	return dep
}

// Execute TODO.
func (f *Taskflow) Execute(ctx context.Context, taskNames ...string) error {
	// validate
	for _, name := range taskNames {
		if !f.isRegistered(name) {
			return fmt.Errorf("%s task was not registered", name) //nolint:goerr113 // TODO
		}
	}

	// run recursive execution
	executedTasks := map[string]bool{}
	for _, name := range taskNames {
		if err := f.execute(ctx, name, executedTasks); err != nil {
			return err
		}
	}
	return nil
}

// MustExecute TODO.
func (f *Taskflow) MustExecute(ctx context.Context, taskNames ...string) {
	err := f.Execute(ctx, taskNames...)
	if err != nil {
		panic(err)
	}
}

// Main TODO.
func (f *Taskflow) Main(args ...string) error {
	ctx := context.Background()
	return f.Execute(ctx, args[1:]...)
}

// execute TODO.
func (f *Taskflow) execute(ctx context.Context, name string, executed map[string]bool) error {
	task := f.tasks[name]
	if executed[name] {
		return nil
	}
	for _, dep := range task.Dependencies {
		if err := f.execute(ctx, dep.name, executed); err != nil {
			return err
		}
	}
	if f.run(ctx, task) {
		return fmt.Errorf("%s task failed", name) //nolint:goerr113 // TODO
	}
	executed[name] = true
	return nil
}

// run TODO.
func (f *Taskflow) run(ctx context.Context, task Task) bool {
	sb := &strings.Builder{}
	tf := &TF{
		ctx:    ctx,
		name:   task.Name,
		writer: sb,
	}

	sb.WriteString(reportTaskStart(task.Name))

	finished := make(chan struct{})
	var duration time.Duration
	go func() {
		defer close(finished)
		from := time.Now()
		task.Command(tf)
		duration = time.Since(from)
	}()
	<-finished

	switch {
	default:
		sb.WriteString(reportTaskEnd("PASS", task.Name, duration))
	case tf.failed:
		sb.WriteString(reportTaskEnd("FAIL", task.Name, duration))
	case tf.skipped:
		sb.WriteString(reportTaskEnd("SKIP", task.Name, duration))
	}

	if f.Verbose || tf.failed {
		if _, err := io.Copy(f.output(), strings.NewReader(sb.String())); err != nil {
			panic(err)
		}
	}

	return tf.failed
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
	return fmt.Sprintf("=== RUN   %s\n", taskName)
}

func reportTaskEnd(status string, taskName string, d time.Duration) string {
	return fmt.Sprintf("--- %s: %s (%.2fs)\n", status, taskName, d.Seconds())
}
