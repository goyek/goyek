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

func TestFlow_trapSignals(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
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
	p.Signal(os.Interrupt)

	// wait for context cancellation
	select {
	case <-ctx.Done():
		// ok
	case <-time.After(time.Second):
		t.Fatal("context should be canceled")
	}

	// second signal
	p.Signal(os.Interrupt)

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
