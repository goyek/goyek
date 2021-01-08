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
	output  io.Writer
	params  Params
	tasks   map[string]Task
	verbose bool
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *flowRunner) Run(ctx context.Context, args []string) int {
	// parse args
	cli := flag.NewFlagSet("", flag.ContinueOnError)
	cli.SetOutput(f.output)
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
		f.verbose = true
	}

	// parse non-flag args
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
			fmt.Fprintf(f.output, "%s task is not registered\n", arg)
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

	_, err := io.WriteString(w, reportTaskStart(task.Name))
	if err != nil {
		panic(err)
	}

	runner := Runner{
		Ctx:     ctx,
		Name:    task.Name,
		Verbose: f.verbose,
		Params:  f.params,
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
		if _, err := io.Copy(f.output, strings.NewReader(sb.String())); err != nil {
			panic(err)
		}
	}

	return !result.failed
}

func reportTaskStart(taskName string) string {
	return fmt.Sprintf("===== TASK  %s\n", taskName)
}

func reportTaskEnd(status string, taskName string, d time.Duration) string {
	return fmt.Sprintf("----- %s: %s (%.2fs)\n", status, taskName, d.Seconds())
}
