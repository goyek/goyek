// Build is the build pipeline for this repository.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

const (
	exitCodeInvalid = 2
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
	skip    = flag.String("skip", "", "do not run actions for given tasks comma-separated names")
	longRun = flag.Duration("long-run", time.Minute, "print when a task takes longer")
)

func main() {
	goyek.SetDefault(all)

	flag.CommandLine.SetOutput(goyek.Output())
	flag.Usage = usage
	flag.Parse()

	if *dryRun {
		*v = true // needed to report the task status
	}

	var skipTasks []string
	if *skip != "" {
		skipTasks = strings.Split(*skip, ",")
	}
	for _, skippedTask := range skipTasks {
		found := false
		for _, task := range goyek.Tasks() {
			if task.Name() == skippedTask {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintln(goyek.Output(), "task to skip is not defined: "+skippedTask)
			usage()
			os.Exit(exitCodeInvalid)
		}
	}

	if *dryRun {
		goyek.Use(middleware.DryRun)
	}
	goyek.Use(middleware.NoRun(skipTasks))
	goyek.Use(middleware.ReportStatus)
	if !*v {
		goyek.Use(middleware.SilentNonFailed)
	}
	if *longRun > 0 {
		goyek.Use(middleware.ReportLongRun(*longRun))
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
