package goyek_test

import (
	"flag"
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func Example() {
	// define a task printing the message (configurable via flag)
	hi := goyek.Define(goyek.Task{
		Name:  "hi",
		Usage: "Greetings",
		Action: func(a *goyek.A) {
			a.Log("Hello world!")
		},
	})

	// define a task running a command
	goVer := goyek.Define(goyek.Task{
		Name:  "go-ver",
		Usage: `Run "go version"`,
		Action: func(a *goyek.A) {
			if err := a.Cmd("go", "version").Run(); err != nil {
				a.Error(err)
			}
		},
	})

	// define a pipeline
	all := goyek.Define(goyek.Task{
		Name: "all",
		Deps: goyek.Deps{hi, goVer},
	})

	// configure middlewares
	goyek.Use(middleware.ReportStatus)

	// set the pipeline as the default task
	goyek.SetDefault(all)

	// run the build pipeline
	goyek.Main(os.Args[1:])

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
	// use the same output for flow and flag
	flag.CommandLine.SetOutput(goyek.Output())

	// define a flag to configure flow output verbosity
	verbose := flag.Bool("v", true, "print all tasks as they are run")

	// define a flag used by a task
	msg := flag.String("msg", "hello world", `message to display by "hi" task`)

	// define a task printing the message (configurable via flag)
	goyek.Define(goyek.Task{
		Name:  "hi",
		Usage: "Greetings",
		Action: func(a *goyek.A) {
			a.Log(*msg)
		},
	})

	// set the help message
	usage := func() {
		fmt.Println("Usage of build: [flags] [--] [tasks]")
		goyek.Print()
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}

	// parse the args
	flag.Usage = usage
	flag.Parse()

	// configure middlewares
	goyek.Use(middleware.ReportStatus)
	if !*verbose {
		goyek.Use(middleware.SilentNonFailed)
	}

	// run the build pipeline
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
