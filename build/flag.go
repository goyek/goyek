package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/goyek/goyek"
)

func run(out io.Writer, flow *goyek.Flow, flags *flag.FlagSet, args []string) {
	flow.Output = out
	flags.SetOutput(out)

	flags.Usage = func() {
		fmt.Fprintln(out, "Usage of build: [flags] [tasks]")
		fmt.Fprintf(out, "Default task: %s\n", flow.DefaultTask.Name())
		w := tabwriter.NewWriter(out, 1, 1, 4, ' ', 0) //nolint:gomnd // ignore

		fmt.Fprintln(out)
		tasks := flow.Tasks()
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name() < tasks[j].Name() })
		fmt.Fprintf(w, "%s\t%s\t%s\n", "Task", "Usage", "Dependencies")
		for _, task := range tasks {
			var deps []string
			for _, dep := range task.Deps() {
				deps = append(deps, dep.Name())
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", task.Name(), task.Usage(), strings.Join(deps, ", "))
		}
		w.Flush() //nolint // not checking errors when writing to output

		fmt.Fprintln(out)
		fmt.Fprintf(w, "%s\t%s\t%s\n", "Flag", "Usage", "Default")
		flags.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "-%s\t%s\t%s\n", f.Name, f.Usage, f.DefValue)
		})
		w.Flush() //nolint // not checking errors when writing to output
	}

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(out, err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	flow.Main(flag.Args())
}
