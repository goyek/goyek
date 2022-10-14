package goyek_test

import (
	"flag"
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
)

func Example() {
	// use the same output for flow and flag
	flow := &goyek.Flow{Output: os.Stdout}
	flag.CommandLine.SetOutput(os.Stdout)

	// define a flag to configure flow verbosity
	flag.BoolVar(&flow.Verbose, "v", true, "print all tasks as they are run")

	// define a flag used by a task
	msg := flag.String("msg", "hello world", `message to display by "hi" task`)

	// define a task printing the message (configurable via flag)
	hi := flow.Define(goyek.Task{
		Name:  "hi",
		Usage: "Greetings",
		Action: func(tf *goyek.TF) {
			tf.Log(*msg)
		},
	})

	// define a task running a command
	goVer := flow.Define(goyek.Task{
		Name:  "go-ver",
		Usage: `Run "go version"`,
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "version").Run(); err != nil {
				tf.Error(err)
			}
		},
	})

	// define a pipeline
	all := flow.Define(goyek.Task{
		Name: "all",
		Deps: goyek.Deps{hi, goVer},
	})

	// set the default task
	flow.SetDefault(all)

	// set the help message
	usage := func() {
		fmt.Println("Usage of build: [flags] [--] [tasks]")
		flow.Print()
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}
	flow.Usage = usage
	flag.Usage = usage

	// parse the args and run the flow
	flag.Parse()
	flow.Main(flag.Args())

	/*
		$ go run . -v
		===== TASK  hi
		      main.go:29: hello world
		----- PASS: hi (0.00s)
		===== TASK  go-ver
		go version go1.19.2 windows/amd64
		----- PASS: go-ver (0.06s)
		ok      0.061s
	*/
}
