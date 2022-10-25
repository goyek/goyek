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
	v      = flag.Bool("v", false, "print all tasks and tests as they are run")
	dryRun = flag.Bool("dry-run", false, "print all tasks that would be run without running them")
)

func main() {
	goyek.SetDefault(all)

	flag.CommandLine.SetOutput(goyek.Output())
	flag.Usage = usage
	flag.Parse()

	if *dryRun {
		*v = true // needed to report the task status
	}

	if *dryRun {
		goyek.Use(middleware.DryRun)
	}
	goyek.Use(middleware.Reporter)
	if !*v {
		goyek.Use(middleware.SilentNonFailed)
	}
	goyek.SetUsage(usage)
	goyek.Main(flag.Args())
}

func usage() {
	fmt.Println("Usage of build: [flags] [--] [tasks]")
	goyek.Print()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}
