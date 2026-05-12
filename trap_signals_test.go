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
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	origTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = origTrapSignalsHook }()
	trapSignalsHookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(trapSignalsHookCalled)
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	var actionCalled bool
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			actionCalled = true
			<-a.Context().Done()
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(doneCh)
	}()

	<-trapSignalsHookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Main to finish")
	}

	if !actionCalled {
		t.Error("action was not called")
	}
	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()
	if gotExitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", gotExitCode, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	origTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = origTrapSignalsHook }()
	trapSignalsHookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(trapSignalsHookCalled)
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			select {} // block forever
		},
	})

	go func() {
		f.Main([]string{"task"})
	}()

	<-trapSignalsHookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// Wait a bit and send second signal
	time.Sleep(100 * time.Millisecond)

	// We need to loop because we want to make sure the second signal is caught
	// by the second select in the signal handler.
	for i := 0; i < 10; i++ {
		mu.Lock()
		gotExitCode := exitCode
		mu.Unlock()
		if gotExitCode == exitCodeFail {
			return
		}
		if err := p.Signal(os.Interrupt); err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatal("Main did not exit with failure after two interrupts")
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	origTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = origTrapSignalsHook }()
	trapSignalsHookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(trapSignalsHookCalled)
	}

	origDefaultFlow := DefaultFlow
	defer func() { DefaultFlow = origDefaultFlow }()
	DefaultFlow = &Flow{}
	DefaultFlow.SetOutput(io.Discard)
	DefaultFlow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})

	doneCh := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(doneCh)
	}()

	<-trapSignalsHookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Main to finish")
	}

	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()
	if gotExitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", gotExitCode, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task"})

	f.Main([]string{"task"})

	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()
	if gotExitCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", gotExitCode, exitCodePass)
	}
}
