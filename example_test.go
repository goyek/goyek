package goyek_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func Example() {
	// Define a task printing a message.
	hi := goyek.Define(goyek.Task{
		Name:  "hi",
		Usage: "Greetings",
		Action: func(a *goyek.A) {
			a.Log("Hello world!")
		},
	})

	// Define a task running a command.
	goVer := goyek.Define(goyek.Task{
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

	// Define a pipeline and set it as the default task.
	all := goyek.Define(goyek.Task{
		Name: "all",
		Deps: goyek.Deps{hi, goVer},
	})
	goyek.SetDefault(all)

	// Configure middlewares.
	goyek.UseExecutor(middleware.ReportFlow)
	goyek.Use(middleware.ReportStatus)
	goyek.Use(middleware.BufferParallel)

	// Run the tasks.
	tasks := os.Args[1:]
	goyek.Main(tasks)

	/*
		$ go run .
		===== TASK  hi
		      main.go:15: Hello world!
		----- PASS: hi (0.00s)
		===== TASK  go-ver
		go version go1.19.2 windows/amd64
		----- PASS: go-ver (0.04s)
		===== TASK  all
		----- NOOP: all (0.00s)
		ok      0.039s
	*/
}

func Example_flag() {
	// Use the same output for flow and flag.
	flag.CommandLine.SetOutput(goyek.Output())

	// Define a flag to configure flow output verbosity.
	verbose := flag.Bool("v", true, "print all tasks as they are run")

	// Define a flag used by a task.
	msg := flag.String("msg", "hello world", `message to display by "hi" task`)

	// Define a task printing the message (configurable via flag).
	goyek.Define(goyek.Task{
		Name:  "hi",
		Usage: "Greetings",
		Action: func(a *goyek.A) {
			a.Log(*msg)
		},
	})

	// Set the help message.
	usage := func() {
		fmt.Println("Usage of build: [flags] [--] [tasks]")
		goyek.Print()
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}

	// Parse the args.
	flag.Usage = usage
	flag.Parse()

	// Configure middlewares.
	goyek.UseExecutor(middleware.ReportFlow)
	goyek.Use(middleware.ReportStatus)
	if *verbose {
		goyek.Use(middleware.BufferParallel)
	} else {
		goyek.Use(middleware.SilentNonFailed)
	}

	// Run the tasks.
	goyek.SetUsage(usage)
	goyek.Main(flag.Args())

	/*
		$ go run .
		no task provided
		Usage of build: [flags] [--] [tasks]
		Tasks:
		   hi   Greetings
		Flags:
		  -msg string
		        message to display by "hi" task (default "hello world")
		  -v    print all tasks as they are run (default true)
		exit status 2
	*/
}
