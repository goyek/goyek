package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
)

const windows = "windows"

func TestFlow_Main_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var mu sync.Mutex
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			p, _ := os.FindProcess(os.Getpid())
			if err := p.Signal(os.Interrupt); err != nil {
				t.Error(err)
			}
			<-a.Context().Done()
		},
	})

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		f.Main([]string{"task"})
	}()
	<-hookCalled

	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var mu sync.Mutex
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			p, _ := os.FindProcess(os.Getpid())
			if err := p.Signal(os.Interrupt); err != nil {
				t.Error(err)
			}
			// Wait for the first signal to be handled
			<-a.Context().Done()

			// Send second signal for hard exit
			if err := p.Signal(os.Interrupt); err != nil {
				t.Error(err)
			}
			// The second signal should trigger osExit(1) in the goroutine
		},
	})

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		f.Main([]string{"task"})
	}()
	<-hookCalled

	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "task"}
	want := "task failed: task"
	if got := err.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
