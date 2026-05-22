package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
)

const (
	windows = "windows"
	task    = "task"
)

func TestFlow_Main_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	// Mock osExit
	var mu sync.Mutex
	var gotExitCode int
	origOsExit := osExit
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		gotExitCode = code
	}
	defer func() { osExit = origOsExit }()

	// Mock hooks
	trapSignalsHookCh := make(chan struct{})
	origTrapSignalsHook := trapSignalsHook
	trapSignalsHook = func() { close(trapSignalsHookCh) }
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	f := &Flow{}
	taskCanFinish := make(chan struct{})
	f.Define(Task{
		Name: task,
		Action: func(a *A) {
			<-a.Context().Done()
			<-taskCanFinish
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{task})
		close(doneCh)
	}()

	<-trapSignalsHookCh
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	close(taskCanFinish)
	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if gotExitCode != exitCodeFail {
		t.Errorf("expected exit code %d, got %d", exitCodeFail, gotExitCode)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	// Mock osExit
	var mu sync.Mutex
	var gotExitCode int
	origOsExit := osExit
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		gotExitCode = code
	}
	defer func() { osExit = origOsExit }()

	// Mock hooks
	trapSignalsHookCh := make(chan struct{})
	origTrapSignalsHook := trapSignalsHook
	trapSignalsHook = func() { close(trapSignalsHookCh) }
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	trapSignalsSecondHookCh := make(chan struct{})
	origTrapSignalsSecondHook := trapSignalsSecondHook
	trapSignalsSecondHook = func() { close(trapSignalsSecondHookCh) }
	defer func() { trapSignalsSecondHook = origTrapSignalsSecondHook }()

	f := &Flow{}
	f.Define(Task{
		Name: task,
		Action: func(a *A) {
			<-a.Context().Done()
			select {} // block forever
		},
	})

	go f.Main([]string{task})

	<-trapSignalsHookCh
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-trapSignalsSecondHookCh
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// Wait for osExit to be called
	for {
		mu.Lock()
		code := gotExitCode
		mu.Unlock()
		if code != 0 {
			break
		}
		runtime.Gosched()
	}

	if gotExitCode != exitCodeFail {
		t.Errorf("expected exit code %d, got %d", exitCodeFail, gotExitCode)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	// Mock osExit
	var mu sync.Mutex
	var gotExitCode int
	origOsExit := osExit
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		gotExitCode = code
	}
	defer func() { osExit = origOsExit }()

	// Mock hooks
	trapSignalsHookCh := make(chan struct{})
	origTrapSignalsHook := trapSignalsHook
	trapSignalsHook = func() { close(trapSignalsHookCh) }
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	origDefaultFlow := DefaultFlow
	DefaultFlow = &Flow{}
	defer func() { DefaultFlow = origDefaultFlow }()

	taskCanFinish := make(chan struct{})
	Define(Task{
		Name: task,
		Action: func(a *A) {
			<-a.Context().Done()
			<-taskCanFinish
		},
	})

	doneCh := make(chan struct{})
	go func() {
		Main([]string{task})
		close(doneCh)
	}()

	<-trapSignalsHookCh
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	close(taskCanFinish)
	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if gotExitCode != exitCodeFail {
		t.Errorf("expected exit code %d, got %d", exitCodeFail, gotExitCode)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	// Mock osExit
	var mu sync.Mutex
	var gotExitCode int
	var exitCalled bool
	origOsExit := osExit
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		gotExitCode = code
		exitCalled = true
	}
	defer func() { osExit = origOsExit }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name:   task,
		Action: func(_ *A) {},
	})

	f.Main([]string{task})

	mu.Lock()
	defer mu.Unlock()
	if !exitCalled {
		t.Fatal("os.Exit was not called")
	}
	if gotExitCode != exitCodePass {
		t.Errorf("expected exit code %d, got %d", exitCodePass, gotExitCode)
	}
}
