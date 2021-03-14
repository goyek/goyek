package taskflow

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type flowRunner struct {
	output      io.Writer
	params      Params
	tasks       map[string]Task
	verbose     bool
	defaultTask RegisteredTask
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *flowRunner) Run(ctx context.Context, args []string) int {
	// prepare flag.FlagSet
	cli := flag.NewFlagSet("", flag.ContinueOnError)
	cli.SetOutput(f.output)
	verbose := cli.Bool("v", false, "Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.")
	usage := func() {
		fmt.Fprintf(cli.Output(), "Usage: [flag(s)] [key=val] task(s)\n")
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
		w.Flush() //nolint // not checking errors when writting to output

		if f.defaultTask.name != "" {
			fmt.Fprintf(cli.Output(), "Default task: %s\n", f.defaultTask.name)
		}

		if len(f.params) > 0 {
			fmt.Fprintf(cli.Output(), "Default parameters:\n")
			for key, val := range f.params {
				fmt.Fprintf(w, "  %s\t%s\n", key, val)
			}
			w.Flush() //nolint // not checking errors when writting to output
		}
	}
	cli.Usage = usage

	// parse args (flags)
	if err := cli.Parse(args); err != nil {
		return CodeInvalidArgs
	}
	if *verbose {
		f.verbose = true
	}

	// parse non-flag args (tasks and parameters)
	var tasks []string
	for _, arg := range cli.Args() {
		if paramAssignmentIdx := strings.IndexRune(arg, '='); paramAssignmentIdx > 0 {
			// parameter assignement via 'key=val'
			key := arg[0:paramAssignmentIdx]
			val := arg[paramAssignmentIdx+1:]
			f.params[key] = val
			continue
		}
		if _, ok := f.tasks[arg]; !ok {
			// task is not registered
			fmt.Fprintf(f.output, "task provided but not registered: %s\n", arg)
			usage()
			return CodeInvalidArgs
		}
		tasks = append(tasks, arg)
	}

	// set default task if none is provided
	if len(tasks) == 0 && f.defaultTask.name != "" {
		tasks = append(tasks, f.defaultTask.name)
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

func (f *flowRunner) run(ctx context.Context, name string, executed map[string]bool) error {
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

func (f *flowRunner) runTask(ctx context.Context, task Task) bool {
	if task.Command == nil {
		return true
	}

	w := f.output
	if !f.verbose {
		w = &strings.Builder{}
	}

	// report task start
	fmt.Fprintf(w, "===== TASK  %s\n", task.Name)

	// run task
	runner := Runner{
		Ctx:      ctx,
		TaskName: task.Name,
		Verbose:  f.verbose,
		Params:   f.params,
		Output:   w,
	}
	result := runner.Run(task.Command)

	// report task end
	status := "PASS"
	switch {
	case result.Failed():
		status = "FAIL"
	case result.Skipped():
		status = "SKIP"
	}
	fmt.Fprintf(w, "----- %s: %s (%.2fs)\n", status, task.Name, result.Duration().Seconds())

	if sb, ok := w.(*strings.Builder); ok && result.failed {
		io.Copy(f.output, strings.NewReader(sb.String())) //nolint // not checking errors when writting to output
	}

	return !result.failed
}
