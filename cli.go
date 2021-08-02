package goyek

import (
	"context"
	"os"
	"os/signal"
)

// Main parses the command-line arguments and runs the provided tasks.
// The usage is printed when invalid arguments are passed.
func (f *Flow) Main() {
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
	}()

	// run flow
	exitCode := f.Run(ctx, os.Args[1:]...)
	os.Exit(exitCode)
}
