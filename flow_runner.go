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
	params      map[string]parameter
	flags       *flag.FlagSet
	tasks       map[string]Task
	verbose     bool
	defaultTask RegisteredTask
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *flowRunner) Run(ctx context.Context, args []string) int {
	// prepare flag.FlagSet
	f.flags = flag.NewFlagSet("", flag.ContinueOnError)
	f.flags.SetOutput(f.output)
	verbose := f.flags.Bool("v", false, "Verbose output: log all tasks as they are run. Also print all text from Log and Logf calls even if the task succeeds.")
	for _, param := range f.params {
		param.register(f.flags)
	}
	usage := func() {
		fmt.Fprintf(f.flags.Output(), "Usage: [flag(s)] task(s)\n")
		fmt.Fprintf(f.flags.Output(), "Flags:\n")
		f.flags.PrintDefaults()

		fmt.Fprintf(f.flags.Output(), "Tasks:\n")
		w := tabwriter.NewWriter(f.flags.Output(), 1, 1, 4, ' ', 0)
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
			params := make([]string, len(t.Parameters))
			for i, param := range t.Parameters {
				params[i] = param.Name()
			}
			sort.Strings(params)
			paramsText := ""
			if len(params) > 0 {
				paramsText = "; -" + strings.Join(params, " -")
			}
			fmt.Fprintf(w, "  %s\t%s%s\n", t.Name, t.Description, paramsText)
		}
		w.Flush() //nolint // not checking errors when writting to output

		if f.defaultTask.name != "" {
			fmt.Fprintf(f.flags.Output(), "Default task: %s\n", f.defaultTask.name)
		}
	}
	f.flags.Usage = usage

	// parse args (flags)
	if err := f.flags.Parse(args); err != nil {
		return CodeInvalidArgs
	}
	if *verbose {
		f.verbose = true
	}

	// parse non-flag args (tasks and parameters)
	var tasks []string
	for _, arg := range f.flags.Args() {
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
		fmt.Fprintln(f.flags.Output(), "no task provided")
		usage()
		return CodeInvalidArgs
	}

	// recursive run
	from := time.Now()
	executedTasks := map[string]bool{}
	for _, name := range tasks {
		if err := f.run(ctx, name, executedTasks); err != nil {
			fmt.Fprintf(f.flags.Output(), "%v\t%.3fs\n", err, time.Since(from).Seconds())
			return CodeFailure
		}
	}
	fmt.Fprintf(f.flags.Output(), "ok\t%.3fs\n", time.Since(from).Seconds())
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

	params := make(map[string]flag.Value)
	for _, param := range task.Parameters {
		params[param.name] = f.flags.Lookup(param.name).Value
	}
	// run task
	runner := Runner{
		Ctx:         ctx,
		TaskName:    task.Name,
		Verbose:     f.verbose,
		ParamValues: params,
		Output:      w,
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
