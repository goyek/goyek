package goyek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type flowRunner struct {
	output      io.Writer
	tasks       map[string]Task
	verbose     bool
	defaultTask RegisteredTask
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *flowRunner) Run(ctx context.Context, args []string) int {
	var tasks []string
	for _, arg := range args {
		if arg == "" {
			fmt.Fprintln(f.output, "the task name cannot be empty")
			return CodeInvalidArgs
		}
		if _, ok := f.tasks[arg]; !ok {
			fmt.Fprintf(f.output, "the task %q is not registred\n", arg)
			return CodeInvalidArgs
		}
		tasks = append(tasks, arg)
	}

	tasks = f.tasksToRun(tasks)

	if len(tasks) == 0 {
		fmt.Fprintln(f.output, "no task provided")
		return CodeInvalidArgs
	}

	return f.runTasks(ctx, tasks)
}

func (f *flowRunner) tasksToRun(tasks []string) []string {
	if len(tasks) > 0 || (f.defaultTask.task.Name == "") {
		return tasks
	}
	return []string{f.defaultTask.task.Name}
}

func (f *flowRunner) runTasks(ctx context.Context, tasks []string) int {
	from := time.Now()
	executedTasks := map[string]bool{}
	for _, name := range tasks {
		if err := f.run(ctx, name, executedTasks); err != nil {
			fmt.Fprintf(f.output, "%v\t%.3fs\n", err, time.Since(from).Seconds())
			return CodeFail
		}
	}
	fmt.Fprintf(f.output, "ok\t%.3fs\n", time.Since(from).Seconds())
	return CodePass
}

func (f *flowRunner) run(ctx context.Context, name string, executed map[string]bool) error {
	task := f.tasks[name]
	if executed[name] {
		return nil
	}
	for _, dep := range task.Deps {
		if err := f.run(ctx, dep.task.Name, executed); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !f.runTask(ctx, task) {
		return errors.New("task failed")
	}
	executed[name] = true
	return nil
}

func (f *flowRunner) runTask(ctx context.Context, task Task) bool {
	if task.Action == nil {
		return true
	}

	writer := f.output
	var streamWriter *strings.Builder
	if !f.verbose {
		streamWriter = &strings.Builder{}
		writer = streamWriter
	}
	writer = &syncWriter{Writer: writer}

	// report task start
	fmt.Fprintf(writer, "===== TASK  %s\n", task.Name)

	tf := &TF{
		ctx:    ctx,
		name:   task.Name,
		output: writer,
	}
	result := tf.run(task.Action)

	// report task end
	status := "PASS"
	passed := true
	switch {
	case result.Failed:
		status = "FAIL"
		passed = false
	case result.Skipped:
		status = "SKIP"
	}
	fmt.Fprintf(writer, "----- %s: %s (%.2fs)\n", status, task.Name, result.Duration.Seconds())

	if streamWriter != nil && result.Failed {
		io.Copy(f.output, strings.NewReader(streamWriter.String())) //nolint // not checking errors when writing to output
	}

	return passed
}

type syncWriter struct {
	io.Writer
	mtx sync.Mutex
}

func (w *syncWriter) Write(p []byte) (int, error) {
	defer func() { w.mtx.Unlock() }()
	w.mtx.Lock()
	return w.Writer.Write(p)
}
