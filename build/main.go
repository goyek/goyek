package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
)

func main() {
	out := os.Stdout
	flow := &goyek.Flow{}
	flags := flag.NewFlagSet("", flag.ExitOnError)

	configure(flow, flags)

	if err := os.Chdir(".."); err != nil { // change working directory to repo root
		fmt.Fprintln(out, err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	run(out, flow, flags, os.Args[1:])
}
