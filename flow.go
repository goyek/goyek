package goyek

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
)

const (
	// CodePass indicates that flow passed.
	CodePass = 0
	// CodeFail indicates that flow failed.
	CodeFail = 1
	// CodeInvalidArgs indicates that flow got invalid input.
	CodeInvalidArgs = 2
)

// Flow is the root type of the package.
// Use Register methods to register all tasks
// and Run or Main method to execute provided tasks.
type Flow struct {
	Output  io.Writer // output where text is printed; os.Stdout by default
	Verbose bool      // control the printing

	DefaultTask RegisteredTask // task which is run when non is explicitly provided

	tasks map[string]Task
}

// Register registers the task. It panics in case of any error.
func (f *Flow) Register(task Task) RegisteredTask {
	// validate
	if task.Name == "" {
		panic("task name cannot be empty")
	}
	if f.isRegistered(task.Name) {
		panic(fmt.Sprintf("%s task was already registered", task.Name))
	}
	for _, dep := range task.Deps {
		if !f.isRegistered(dep.task.Name) {
			panic(fmt.Sprintf("invalid dependency %s", dep.task.Name))
		}
	}

	f.tasks[task.Name] = task
	return RegisteredTask{task: task}
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *Flow) Run(ctx context.Context, args ...string) int {
	if ctx == nil {
		ctx = context.Background()
	}

	flow := &flowRunner{
		output:      f.Output,
		tasks:       f.tasks,
		verbose:     f.Verbose,
		defaultTask: f.DefaultTask,
	}

	if flow.output == nil {
		flow.output = os.Stdout
	}

	return flow.Run(ctx, args)
}

// Main parses the args and runs the provided tasks.
// It exists when after the taskflow finished or SIGINT
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
func (f *Flow) Tasks() []RegisteredTask {
	var tasks []RegisteredTask
	for _, task := range f.tasks {
		tasks = append(tasks, RegisteredTask{
			task: task,
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name() < tasks[j].Name() })
	return tasks
}

// TODO: Print (like PrintDefaults)

func (f *Flow) isRegistered(name string) bool {
	if f.tasks == nil {
		f.tasks = map[string]Task{}
	}
	_, ok := f.tasks[name]
	return ok
}
