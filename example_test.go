package goyek_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

var (
	v      = flag.Bool("v", false, "print all tasks and tests as they are run")
	dryRun = flag.Bool("dry-run", false, "print all tasks that would be run without running them")
	noDeps = flag.Bool("no-deps", false, "do not process dependencies")
	skip   = flag.String("skip", "", "skip processing the `comma-separated tasks`")
	msg    = flag.String("msg", "hello world", `message to display by "hi" task`)
)

var all = goyek.Define(goyek.Task{
	Name: "all",
	Deps: goyek.Deps{hi, goVer},
})

var hi = goyek.Define(goyek.Task{
	Name:  "hi",
	Usage: "Greetings",
	Action: func(a *goyek.A) {
		a.Log(*msg)
	},
})

var goVer = goyek.Define(goyek.Task{
	Name:  "go-ver",
	Usage: `Run "go version"`,
	Action: func(a *goyek.A) {
		cmd := exec.CommandContext(a.Context(), "go", "version")
		cmd.Stdout = a.Output()
		cmd.Stderr = a.Output()
		if err := cmd.Run(); err != nil {
			a.Error(err)
		}
	},
})

func Example() {
	// Use the same output for flow and flag.
	flag.CommandLine.SetOutput(goyek.Output())

	// Set the help message.
	usage := func() {
		fmt.Println("Usage of build: [tasks] [flags] [--] [args]")
		goyek.Print()
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}

	// Parse the args.
	flag.Usage = usage
	tasks, err := goyek.Parse(nil, nil)
	if err != nil {
		fmt.Fprintln(goyek.Output(), err)
		os.Exit(2)
	}

	// Configure middlewares.
	if *dryRun {
		goyek.Use(middleware.DryRun)
	}
	goyek.Use(middleware.ReportStatus)
	if *v {
		goyek.Use(middleware.BufferParallel)
	} else {
		goyek.Use(middleware.SilentNonFailed)
	}

	var opts []goyek.Option
	if *noDeps {
		opts = append(opts, goyek.NoDeps())
	}
	if *skip != "" {
		skippedTasks := strings.Split(*skip, ",")
		opts = append(opts, goyek.Skip(skippedTasks...))
	}

	// Run the tasks.
	goyek.SetDefault(all)
	goyek.SetUsage(usage)
	if err := goyek.Execute(context.Background(), tasks, opts...); err != nil {
		fmt.Println(err)
	}
}
