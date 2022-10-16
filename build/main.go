package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

var flow = &goyek.Flow{}

func main() {
	// change working directory to repo root
	if err := os.Chdir(".."); err != nil {
		fmt.Println(err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	// flags
	verbose := flag.Bool("v", false, "print all tasks as they are run")
	ci := flag.Bool("ci", false, "whether CI is calling")

	// tasks
	clean := flow.Define(taskClean())
	modTidy := flow.Define(taskModTidy())
	install := flow.Define(taskInstall())
	build := flow.Define(taskBuild())
	markdownlint := flow.Define(taskMarkdownLint())
	misspell := flow.Define(taskMisspell())
	golangciLint := flow.Define(taskGolangciLint())
	test := flow.Define(taskTest())
	diff := flow.Define(taskDiff(ci))

	// pipelines
	lint := flow.Define(taskLint(goyek.Deps{
		misspell,
		markdownlint,
		golangciLint,
	}))
	all := flow.Define(taskAll(goyek.Deps{
		clean,
		modTidy,
		install,
		build,
		lint,
		test,
		diff,
	}))

	// set default task
	flow.SetDefault(all)

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

	// configure middlewares
	flow.Use(middleware.Reporter)
	if !*verbose {
		flow.Use(middleware.SilentNonFailed)
	}

	flow.Main(flag.Args())
}
