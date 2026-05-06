package goyek

import (
	"context"
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestFlow_Main(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}

	f := &Flow{}
	f.Define(Task{
		Name: "test",
		Action: func(a *A) {
			// wait for signal
			select {
			case <-a.Context().Done():
			case <-time.After(time.Second):
				t.Error("timeout waiting for context cancellation")
			}
		},
	})

	go func() {
		// wait for Main to start and register signal handler
		time.Sleep(100 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		if err := p.Signal(os.Interrupt); err != nil {
			t.Errorf("failed to send signal: %v", err)
		}
	}()

	f.Main([]string{"test"})

	if atomic.LoadInt32(&exitCode) != 1 {
		t.Errorf("expected exit code 1, got %d", atomic.LoadInt32(&exitCode))
	}
}

func TestFlow_trapSignals(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}

	f := &Flow{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	go f.trapSignals(ctx, cancel, io.Discard, done)

	p, _ := os.FindProcess(os.Getpid())

	// first signal
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("failed to send first signal: %v", err)
	}

	// wait for context cancellation
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for context cancellation")
	}

	// second signal
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("failed to send second signal: %v", err)
	}

	// wait for hard exit
	for i := 0; i < 10; i++ {
		if atomic.LoadInt32(&exitCode) == 1 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if atomic.LoadInt32(&exitCode) != 1 {
		t.Errorf("expected exit code 1, got %d", atomic.LoadInt32(&exitCode))
	}

	close(done)
}

func TestMain_topLevel(t *testing.T) {
	// Just to cover the top-level Main function
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	osExit = func(code int) {}

	// Define a task in DefaultFlow
	f := DefaultFlow
	if f.tasks == nil {
		f.tasks = make(map[string]*DefinedTask)
	}
	f.Define(Task{Name: "top", Action: func(a *A) {}})
	Main([]string{"top"})
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "test"}
	if err.Error() != "task failed: test" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}
