// Build is the build pipeline for this repository.
package main

import (
	"flag"
	"fmt"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

// Directories used in repository.
const (
	dirRoot  = "."
	dirBuild = "build"
)

// Reusable flags used by the build pipeline.
var (
	v = flag.Bool("v", false, "print all tasks and tests as they are run")
)

var flow = &goyek.Flow{}

func main() {
	flag.CommandLine.SetOutput(flow.Output())
	flag.Usage = usage
	flag.Parse()

	flow.SetDefault(all)
	flow.Use(middleware.Reporter)
	if !*v {
		flow.Use(middleware.SilentNonFailed)
	}
	flow.SetUsage(usage)
	flow.Main(flag.Args())
}

func usage() {
	fmt.Println("Usage of build: [flags] [--] [tasks]")
	flow.Print()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}
