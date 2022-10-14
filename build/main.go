package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
)

var flow = &goyek.Flow{}

func main() {
	// change working directory to repo root
	if err := os.Chdir(".."); err != nil {
		fmt.Println(err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	configure()

	flow.Output = os.Stdout
	flag.CommandLine.SetOutput(os.Stdout)
	usage := func() {
		fmt.Println("Usage of build: [flags] [--] [tasks]")
		flow.Print()
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}
	flow.Usage = usage
	flag.Usage = usage

	flag.Parse()
	flow.Main(flag.Args())
}
