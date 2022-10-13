package goyek

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

// Main parses the command-line arguments and runs the provided tasks.
// The usage is printed when invalid arguments are passed.
func (f *Flow) Main(args []string) {
	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
		<-c // second signal, hard exit
		fmt.Fprintln(f.Output, "second interrupt signal, hard exit")
		os.Exit(CodeFail)
	}()

	// run flow
	exitCode := f.Run(ctx, args...)
	os.Exit(exitCode)
}
