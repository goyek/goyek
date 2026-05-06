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

	// First signal (not Interrupt)
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

func TestFlow_handleSignals_interrupt(t *testing.T) {
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

	// First signal (Interrupt)
	c <- os.Interrupt

	select {
	case <-ctx.Done():
		// Success: context canceled
	case <-time.After(time.Second):
		t.Fatal("context not canceled after first signal")
	}

	// Second signal
	c <- os.Interrupt

	select {
	case code := <-exitCalled:
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	case <-time.After(time.Second):
		t.Fatal("exit not called after second signal")
	}
}

func TestFlow_Main_exported(t *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	exitCalled := make(chan int, 1)
	osExit = func(code int) {
		exitCalled <- code
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	// Test (f *Flow).Main
	flow.Main([]string{"task"})
	if got := <-exitCalled; got != 0 {
		t.Errorf("expected exit code 0, got %d", got)
	}

	// Test Main
	flowBackup := DefaultFlow
	DefaultFlow = flow
	defer func() { DefaultFlow = flowBackup }()

	Main([]string{"task"})
	if got := <-exitCalled; got != 0 {
		t.Errorf("expected exit code 0, got %d", got)
	}

	// Test Main with invalid args
	Main([]string{"invalid"})
	if got := <-exitCalled; got != 2 {
		t.Errorf("expected exit code 2, got %d", got)
	}
}

type stringsBuilder struct {
	io.Writer
}

func (b *stringsBuilder) Write(p []byte) (n int, err error) {
	return len(p), nil
}
