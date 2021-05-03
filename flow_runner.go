package goyek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type flowRunner struct {
	output      io.Writer
	params      map[string]paramValueFactory
	paramValues map[string]ParamValue
	tasks       map[string]Task
	verbose     RegisteredBoolParam
	defaultTask RegisteredTask
}

// Run runs provided tasks and all their dependencies.
// Each task is executed at most once.
func (f *flowRunner) Run(ctx context.Context, args []string) int { //nolint // TODO: refactor
	if unusedParams := f.unusedParams(); len(unusedParams) > 0 {
		panic(fmt.Sprintf("unused parameters: %v\n", unusedParams))
	}

	f.paramValues = make(map[string]ParamValue)
	for _, param := range f.params {
		value := param.newValue()
		f.paramValues[param.name] = value
	}
	usageRequested := false

	var argHandler func(string) error

	handleNextArgFor := func(value ParamValue) {
		nextHandler := argHandler
		argHandler = func(s string) error {
			err := value.Set(s)
			argHandler = nextHandler
			return err
		}
	}

	var tasks []string

	argHandler = func(arg string) error {
		if _, isTask := f.tasks[arg]; isTask {
			tasks = append(tasks, arg)
			return nil
		}
		if arg[0] == '-' {
			// parse parameters
			split := strings.SplitN(arg[1:], "=", 2)
			if value, isFlag := f.paramValues[split[0]]; isFlag {
				switch {
				case len(split) > 1:
					return value.Set(split[1])
				case value.IsBool():
					return value.Set("")
				default:
					handleNextArgFor(value)
					return nil
				}
			}
		}
		// if they haven't been overridden above, provide usage for common queries
		if (arg == "-h") || (arg == "--help") || (arg == "help") {
			usageRequested = true
			return nil
		}
		fmt.Fprintf(f.output, "unknown argument: %s\n", arg)
		return fmt.Errorf("unknown argument: %s", arg)
	}

	for _, arg := range args {
		err := argHandler(arg)
		if err != nil {
			return CodeInvalidArgs
		}
	}

	if usageRequested {
		printUsage(f)
		return CodePass
	}

	// set default task if none is provided
	if len(tasks) == 0 && f.defaultTask.name != "" {
		tasks = append(tasks, f.defaultTask.name)
	}

	if len(tasks) == 0 {
		fmt.Fprintln(f.output, "no task provided")
		printUsage(f)
		return CodeInvalidArgs
	}

	return f.runTasks(ctx, tasks)
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

	paramValues := make(map[string]ParamValue)
	for _, param := range task.Params {
		paramValues[param.Name()] = f.paramValues[param.Name()]
	}

	// if verbose flag is registered then check its value
	verboseParamVal, ok := f.paramValues[f.verbose.Name()]
	verbose := ok && verboseParamVal.Get().(bool)

	failed := false
	measuredCommand := func(tf *TF) {
		w := tf.Output()
		if !verbose {
			w = &strings.Builder{}
		}

		// report task start
		fmt.Fprintf(w, "===== TASK  %s\n", tf.Name())

		// run task
		r := runner{
			Ctx:         tf.Context(),
			TaskName:    tf.Name(),
			ParamValues: tf.paramValues,
			Output:      w,
		}
		result := r.Run(task.Command)

		// report task end
		status := "PASS"
		switch {
		case result.Failed():
			status = "FAIL"
			failed = true
		case result.Skipped():
			status = "SKIP"
		}
		fmt.Fprintf(w, "----- %s: %s (%.2fs)\n", status, tf.Name(), result.Duration().Seconds())

		if sb, ok := w.(*strings.Builder); ok && result.failed {
			io.Copy(tf.Output(), strings.NewReader(sb.String())) //nolint // not checking errors when writting to output
		}
	}

	measuredRunner := runner{
		Ctx:         ctx,
		TaskName:    task.Name,
		ParamValues: paramValues,
		Output:      f.output,
	}
	measuredRunner.Run(measuredCommand)

	return !failed
}

func (f *flowRunner) unusedParams() []string {
	remainingParams := make(map[string]struct{})
	for key := range f.params {
		remainingParams[key] = struct{}{}
	}
	delete(remainingParams, f.verbose.Name())
	for _, task := range f.tasks {
		for _, param := range task.Params {
			delete(remainingParams, param.Name())
		}
	}
	unusedParams := make([]string, 0, len(remainingParams))
	for key := range remainingParams {
		unusedParams = append(unusedParams, key)
	}
	return unusedParams
}

func printUsage(f *flowRunner) {
	flagName := func(paramName string) string {
		return "-" + paramName
	}

	fmt.Fprintf(f.output, "Usage: [flag(s) | task(s)]...\n")
	fmt.Fprintf(f.output, "Flags:\n")
	w := tabwriter.NewWriter(f.output, 1, 1, 4, ' ', 0)
	keys := make([]string, 0, len(f.params))
	for key := range f.params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		param := f.params[key]
		fmt.Fprintf(w, "  %s\tDefault: %s\t%s\n", flagName(param.name), param.newValue().String(), param.usage)
	}
	w.Flush() //nolint // not checking errors when writing to output

	fmt.Fprintf(f.output, "Tasks:\n")
	keys = make([]string, 0, len(f.tasks))
	for k, task := range f.tasks {
		if task.Usage == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t := f.tasks[k]
		params := make([]string, len(t.Params))
		for i, param := range t.Params {
			params[i] = flagName(param.Name())
		}
		sort.Strings(params)
		paramsText := ""
		if len(params) > 0 {
			paramsText = "; " + strings.Join(params, " ")
		}
		fmt.Fprintf(w, "  %s\t%s%s\n", t.Name, t.Usage, paramsText)
	}
	w.Flush() //nolint // not checking errors when writing to output

	if f.defaultTask.name != "" {
		fmt.Fprintf(f.output, "Default task: %s\n", f.defaultTask.name)
	}
}
