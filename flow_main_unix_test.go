//go:build !windows && !plan9
// +build !windows,!plan9

package goyek

import (
	"context"
	"io"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestFlow_handleSignals(t *testing.T) {
	flow := &Flow{}
	out := &stringsBuilder{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 2)

	exitCalled := make(chan int, 1)
	exitFunc := func(code int) {
		exitCalled <- code
	}

	go flow.handleSignals(c, out, cancel, exitFunc)

	// First signal
	c <- syscall.SIGTERM

	select {
	case <-ctx.Done():
		// Success: context canceled
	case <-time.After(time.Second):
		t.Fatal("context not canceled after first signal")
	}

	// Second signal
	c <- syscall.SIGTERM

	select {
	case code := <-exitCalled:
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	case <-time.After(time.Second):
		t.Fatal("exit not called after second signal")
	}
}

type stringsBuilder struct {
	io.Writer
}

func (b *stringsBuilder) Write(p []byte) (n int, err error) {
	return len(p), nil
}
