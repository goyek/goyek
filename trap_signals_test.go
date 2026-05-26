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

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: task,
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{task})
		close(done)
	}()

	p, _ := os.FindProcess(os.Getpid())
	for {
		_ = p.Signal(os.Interrupt)
		runtime.Gosched()
		select {
		case <-hookCalled:
			goto graceful
		default:
		}
	}
graceful:
	close(taskCanFinish)
	<-done

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	origTrapSignalsSecondHook := trapSignalsSecondHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		trapSignalsSecondHook = origTrapSignalsSecondHook
	}()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}
	secondHookCalled := make(chan struct{})
	trapSignalsSecondHook = func() {
		close(secondHookCalled)
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{
		Name: task,
		Action: func(_ *A) {
			select {} // block forever
		},
	})

	go flow.Main([]string{task})

	p, _ := os.FindProcess(os.Getpid())

	for {
		_ = p.Signal(os.Interrupt)
		runtime.Gosched()
		select {
		case <-hookCalled:
			goto secondSignal
		default:
		}
	}

secondSignal:
	for {
		_ = p.Signal(os.Interrupt)
		runtime.Gosched()
		select {
		case <-secondHookCalled:
			goto finished
		default:
		}
	}

finished:

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	origDefaultFlow := DefaultFlow
	defer func() { DefaultFlow = origDefaultFlow }()
	DefaultFlow = &Flow{}
	DefaultFlow.SetOutput(io.Discard)
	taskCanFinish := make(chan struct{})
	DefaultFlow.Define(Task{
		Name: task,
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		Main([]string{task})
		close(done)
	}()

	p, _ := os.FindProcess(os.Getpid())
	for {
		_ = p.Signal(os.Interrupt)
		runtime.Gosched()
		select {
		case <-hookCalled:
			goto gracefulMain
		default:
		}
	}
gracefulMain:
	close(taskCanFinish)
	<-done

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: task})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{task})
		close(done)
	}()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}
