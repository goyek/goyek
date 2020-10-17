package taskflow

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	// ErrTaskNotRegistered TODO.
	ErrTaskNotRegistered = errors.New("task provided but not registered")
	// ErrTaskFail TODO.
	ErrTaskFail = errors.New("FAIL")
)

// Taskflow TODO.
type Taskflow struct {
	Verbose bool
	Output  io.Writer

	tasks map[string]Task
}

// RegisteredTask TODO.
type RegisteredTask struct {
	name string
}

// Register TODO.
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

// MustRegister TODO.
func (f *Taskflow) MustRegister(task Task) RegisteredTask {
	dep, err := f.Register(task)
	if err != nil {
		panic(err)
	}
	return dep
}

// Main TODO.
func (f *Taskflow) Main(args ...string) {
	ctx := context.Background()
	cli := flag.NewFlagSet("", flag.ExitOnError)
	cli.SetOutput(f.Output)
	verbose := cli.Bool("v", false, "verbose")
	usage := func() {
		fmt.Fprintf(cli.Output(), "Usage: [flag(s)] task(s)\n")
		fmt.Fprintf(cli.Output(), "Flags:\n")
		cli.PrintDefaults()
		fmt.Fprintf(cli.Output(), "Tasks:\n")
		for _, t := range f.tasks {
			fmt.Fprintf(cli.Output(), "  %s\t%s\n", t.Name, t.Description)
		}
	}
	cli.Usage = usage
	cli.Parse(args[1:]) //nolint // Ignore errors; FlagSet is set for ExitOnError.
	if *verbose {
		f.Verbose = true
	}
	if err := f.Execute(ctx, cli.Args()...); err != nil {
		fmt.Fprintln(cli.Output(), err)
		if err == ErrTaskNotRegistered {
			usage()
		}
		os.Exit(1)
	}
}

// Execute TODO.
func (f *Taskflow) Execute(ctx context.Context, taskNames ...string) error {
	// validate
	for _, name := range taskNames {
		if !f.isRegistered(name) {
			return ErrTaskNotRegistered
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
		return ErrTaskFail
	}
	executed[name] = true
	return nil
}

// run TODO.
func (f *Taskflow) run(ctx context.Context, task Task) bool {
	// 1. Handle cancelation via ctx. New state? Check how go test does it. TODO.
	// 2. Handle writer streaming for verbose mode.
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
	return fmt.Sprintf("===== TASK  %s\n", taskName)
}

func reportTaskEnd(status string, taskName string, d time.Duration) string {
	return fmt.Sprintf("----- %s: %s (%.2fs)\n", status, taskName, d.Seconds())
}
