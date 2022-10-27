// Build is the build pipeline for this repository.
package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

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
	v       = flag.Bool("v", false, "print all tasks and tests as they are run")
	dryRun  = flag.Bool("dry-run", false, "print all tasks that would be run without running them")
	longRun = flag.Duration("long-run", time.Minute, "print when a task takes longer")
	noDeps  = flag.Bool("no-deps", false, "do not process dependencies")
	skip    = flag.String("skip", "", "skip processing the `comma-separated tasks`")
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
	goyek.Use(middleware.ReportStatus)
	if !*v {
		goyek.Use(middleware.SilentNonFailed)
	}
	if *longRun > 0 {
		goyek.Use(middleware.ReportLongRun(*longRun))
	}

	var opts []goyek.Option
	if *noDeps {
		opts = append(opts, goyek.NoDeps())
	}
	if *skip != "" {
		skippedTasks := strings.Split(*skip, ",")
		opts = append(opts, goyek.Skip(skippedTasks...))
	}

	goyek.SetUsage(usage)
	goyek.Main(flag.Args(), opts...)
}

func usage() {
	fmt.Println("Usage of build: [flags] [--] [tasks]")
	goyek.Print()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}
