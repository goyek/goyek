package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
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

func TestMain_signal_graceful(t *testing.T) {
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
	task := f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			p, _ := os.FindProcess(os.Getpid())
			if err := p.Signal(os.Interrupt); err != nil {
				t.Error(err)
			}
			<-a.Context().Done()
		},
	})
	f.SetDefault(task)

	origDefaultFlow := DefaultFlow
	DefaultFlow = f
	defer func() {
		DefaultFlow = origDefaultFlow
	}()

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		Main(nil)
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
			// Send first signal
			if err := p.Signal(os.Interrupt); err != nil {
				t.Error(err)
			}
			// Wait for the first signal to be handled
			<-a.Context().Done()

			// Send second signal for hard exit
			// Use a loop to increase the chance of it being caught by the second select
			for i := 0; i < 100; i++ {
				_ = p.Signal(os.Interrupt)
				time.Sleep(10 * time.Millisecond)
				mu.Lock()
				done := exitCode != 0
				mu.Unlock()
				if done {
					return
				}
			}
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

func TestFlow_Main_pass(t *testing.T) {
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
	if exitCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodePass)
	}
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "task"}
	want := "task failed: task"
	if got := err.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
