package goyek

import (
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

const windows = "windows"

func TestFlow_Main_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}

	oldOsExit := osExit
	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}
	defer func() { osExit = oldOsExit }()

	f := &Flow{}
	f.SetOutput(io.Discard)

	f.Define(Task{
		Name: "sleep",
		Action: func(a *A) {
			p, _ := os.FindProcess(os.Getpid())
			if err := p.Signal(os.Interrupt); err != nil {
				t.Fatal(err)
			}

			select {
			case <-a.Context().Done():
			case <-time.After(5 * time.Second):
				a.Error("should have been canceled")
			}
		},
	})

	f.Main([]string{"sleep"})

	if got := atomic.LoadInt32(&exitCode); got != 1 {
		t.Errorf("expected exit code 1, got %d", got)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}

	oldOsExit := osExit
	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}
	defer func() { osExit = oldOsExit }()

	f := &Flow{}
	f.SetOutput(io.Discard)

	f.Define(Task{
		Name: "sleep",
		Action: func(a *A) {
			p, _ := os.FindProcess(os.Getpid())
			if err := p.Signal(os.Interrupt); err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)
			if err := p.Signal(os.Interrupt); err != nil {
				t.Fatal(err)
			}

			select {
			case <-a.Context().Done():
			case <-time.After(5 * time.Second):
				a.Error("should have been canceled")
			}
		},
	})

	f.Main([]string{"sleep"})

	if got := atomic.LoadInt32(&exitCode); got != 1 {
		t.Errorf("expected exit code 1, got %d", got)
	}
}

func TestMain_topLevel(t *testing.T) {
	// Just to cover the top-level Main wrapper
	oldOsExit := osExit
	osExit = func(code int) {}
	defer func() { osExit = oldOsExit }()

	Main(nil)
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "foo"}
	want := "task failed: foo"
	if got := err.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
