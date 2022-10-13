package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/goyek/goyek"
)

func run(out io.Writer, flow *goyek.Flow, flags *flag.FlagSet, args []string) {
	flow.Output = out
	flags.SetOutput(out)

	flags.Usage = func() {
		fmt.Fprintln(out, "Usage of build: [flags] [--] [tasks]")
		flow.Print()
		fmt.Fprintln(out, "Flags:")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(out, err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	flow.Main(flags.Args())
}
