// Build is the build pipeline for this repository.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

// Directories used in repository.
const (
	dirRoot  = "."
	dirBuild = "build"
	dirTools = "tools"
)

// Reusable flags used by the build pipeline.
var (
	v = flag.Bool("v", false, "print all tasks and tests as they are run")
)

var flow = &goyek.Flow{}

func main() {
	flow.SetDefault(all)

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

	flow.Use(middleware.Reporter)
	if !*v {
		flow.Use(middleware.SilentNonFailed)
	}

	flow.Main(flag.Args())
}
