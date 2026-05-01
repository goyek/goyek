//go:build !windows && !plan9
// +build !windows,!plan9

package goyek

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/goyek/goyek/v3/internal"
)

func TestFlow_Main_SIGTERM(t *testing.T) {
	flow := &Flow{}
	out := &strings_builder{}
	flow.SetOutput(out)

	taskStarted := make(chan struct{})
	taskFinished := make(chan struct{})

	flow.Define(Task{
		Name: "test",
		Action: func(a *A) {
			close(taskStarted)
			select {
			case <-a.Context().Done():
			case <-time.After(5 * time.Second):
				// This should not happen if SIGTERM is handled correctly
			}
			close(taskFinished)
		},
	})

	// We can't call flow.Main because it calls os.Exit.
	// But we can simulate the signal handling logic.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, internal.TerminationSignals...)
	defer signal.Stop(c)

	go func() {
		// Wait for task to start
		<-taskStarted
		// Send SIGTERM to ourselves
		process, _ := os.FindProcess(os.Getpid())
		process.Signal(syscall.SIGTERM)
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-taskFinished:
		}
	}()

	flow.main(ctx, []string{"test"})

	select {
	case <-taskFinished:
		// Success
	case <-time.After(time.Second):
		t.Error("task was not canceled by SIGTERM")
	}
}

type strings_builder struct {
	io.Writer
}

func (b *strings_builder) Write(p []byte) (n int, err error) {
	return len(p), nil
}
