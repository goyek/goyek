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

func TestFlow_trapSignals(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows because os.Process.Signal(os.Interrupt) is not supported")
	}
	f := &Flow{}
	f.SetOutput(io.Discard)
	ctx, cancel := context.WithCancel(context.Background())

	var exitCode int32 = -1
	origOsExit := osExit
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}
	defer func() { osExit = origOsExit }()

	f.trapSignals(cancel)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("could find process: %v", err)
	}

	// Send first signal.
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("could not send signal: %v", err)
	}

	select {
	case <-ctx.Done():
		// Success
	case <-time.After(time.Second):
		t.Error("context should have been canceled")
	}

	// Send second signal.
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("could not send signal: %v", err)
	}

	// Wait for osExit to be called.
	start := time.Now()
	for time.Since(start) < time.Second {
		if atomic.LoadInt32(&exitCode) != -1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got := atomic.LoadInt32(&exitCode); got != exitCodeFail {
		t.Errorf("got exit code: %d; want: %d", got, exitCodeFail)
	}
}

func TestFlow_Main(t *testing.T) {
	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task"})

	var exitCode int32 = -1
	origOsExit := osExit
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}
	defer func() { osExit = origOsExit }()

	f.Main([]string{"task"})

	if got := atomic.LoadInt32(&exitCode); got != exitCodePass {
		t.Errorf("got exit code: %d; want: %d", got, exitCodePass)
	}
}

func TestMain_topLevel(t *testing.T) {
	Define(Task{Name: "top-level-task"})

	var exitCode int32 = -1
	origOsExit := osExit
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}
	defer func() { osExit = origOsExit }()

	Main([]string{"top-level-task"})

	if got := atomic.LoadInt32(&exitCode); got != exitCodePass {
		t.Errorf("got exit code: %d; want: %d", got, exitCodePass)
	}
}
