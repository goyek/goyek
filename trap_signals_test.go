package goyek

import (
	"context"
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

const windows = "windows"

func TestFlow_trapSignals(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on " + windows)
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	f := &Flow{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	defer close(done)
	f.trapSignals(ctx, cancel, io.Discard, done)

	p, _ := os.FindProcess(os.Getpid())

	// first signal
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// wait for context cancellation
	select {
	case <-ctx.Done():
		// ok
	case <-time.After(time.Second):
		t.Fatal("context should be canceled")
	}

	// second signal
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// wait for osExit call
	start := time.Now()
	for {
		mu.Lock()
		code := exitCode
		mu.Unlock()
		if code == exitCodeFail {
			break
		}
		if time.Since(start) > time.Second {
			t.Fatal("osExit should be called with exitCodeFail")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestFlow_Main(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	f := &Flow{}
	f.Main(nil)

	if exitCode != exitCodeInvalid {
		t.Errorf("expected exitCodeInvalid, got %d", exitCode)
	}
}

func TestMain_topLevel(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	Main(nil)

	if exitCode != exitCodeInvalid {
		t.Errorf("expected exitCodeInvalid, got %d", exitCode)
	}
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "task1"}
	got := err.Error()
	want := "task failed: task1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFlow_trapSignals_done(_ *testing.T) {
	f := &Flow{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	close(done)
	f.trapSignals(ctx, cancel, io.Discard, done)
}

func TestFlow_trapSignals_ctxDone(_ *testing.T) {
	f := &Flow{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	defer close(done)
	f.trapSignals(ctx, cancel, io.Discard, done)
}

func TestFlow_trapSignals_done_second(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on " + windows)
	}
	f := &Flow{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	f.trapSignals(ctx, cancel, io.Discard, done)

	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ctx.Done():
		// ok
	case <-time.After(time.Second):
		t.Fatal("context should be canceled")
	}

	close(done)
}
