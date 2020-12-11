package taskflow

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"text/tabwriter"
	"time"
)

// Main parses the command-line arguments and runs the provided tasks.
// The usage is printed when invalid arguments are passed.
func (f *Taskflow) Main() {
	// parse args
	cli := flag.NewFlagSet("", flag.ExitOnError)
	cli.SetOutput(f.Output)
	verbose := cli.Bool("v", false, "verbose")
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
	cli.Parse(os.Args[1:]) //nolint // Ignore errors; FlagSet is set for ExitOnError.
	tasks := cli.Args()
	if len(tasks) == 0 {
		fmt.Fprintln(cli.Output(), "no task provided")
		usage()
		os.Exit(1)
	}

	if *verbose {
		f.Verbose = true
	}

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	// run tasks
	from := time.Now()
	err := f.Run(ctx, tasks...)
	duration := time.Since(from)
	if err != nil {
		fmt.Fprintf(cli.Output(), "%v\t%.3fs\n", err, duration.Seconds())
		if err == ErrTaskNotRegistered {
			usage()
		}
		os.Exit(1)
	}
	fmt.Fprintf(cli.Output(), "ok\t%.3fs\n", duration.Seconds())
}
