package goyek

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTrapTerminationSignalsDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	signals := make(chan os.Signal)
	exited := make(chan int, 1)
	var out strings.Builder
	handlerDone := trapTerminationSignals(&out, signals, done, cancel, func(code int) {
		exited <- code
	})

	close(done)
	waitForDone(t, handlerDone)
	assertNoExit(t, exited)

	if err := ctx.Err(); err != nil {
		t.Fatalf("context error: %v", err)
	}
	if got := out.String(); got != "" {
		t.Fatalf("got output %q, want empty", got)
	}
}

func TestTrapTerminationSignalsGraceful(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	signals := make(chan os.Signal, 1)
	exited := make(chan int, 1)
	var out strings.Builder
	handlerDone := trapTerminationSignals(&out, signals, done, cancel, func(code int) {
		exited <- code
	})

	signals <- os.Interrupt
	waitForContext(t, ctx)
	close(done)
	waitForDone(t, handlerDone)
	assertNoExit(t, exited)

	if got, want := out.String(), "first interrupt, graceful stop\n"; got != want {
		t.Fatalf("got output %q, want %q", got, want)
	}
}

func TestTrapTerminationSignalsHardExit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	signals := make(chan os.Signal, 1)
	var out strings.Builder
	exitCode := -1
	handlerDone := trapTerminationSignals(&out, signals, done, cancel, func(code int) {
		exitCode = code
	})

	signals <- os.Interrupt
	waitForContext(t, ctx)
	signals <- os.Interrupt
	waitForDone(t, handlerDone)
	close(done)

	if exitCode != exitCodeFail {
		t.Fatalf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
	want := "first interrupt, graceful stop\nsecond interrupt, exit\n"
	if got := out.String(); got != want {
		t.Fatalf("got output %q, want %q", got, want)
	}
}

func waitForContext(t *testing.T, ctx context.Context) {
	t.Helper()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("context should have been canceled")
	}
}

func waitForDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for goroutine")
	}
}

func assertNoExit(t *testing.T, exited <-chan int) {
	t.Helper()
	select {
	case code := <-exited:
		t.Fatalf("exit called with code %d", code)
	default:
	}
}
