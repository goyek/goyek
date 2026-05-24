package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
)

const (
	windows = "windows"
	task    = "task"
)

func TestFlow_Main_signal(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on " + windows)
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var mu sync.Mutex
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	origTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = origTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	f := &Flow{}
	taskCanFinish := make(chan struct{})
	f.Define(Task{
		Name: task,
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		f.Main([]string{task})
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)

	// Wait a bit to ensure the signal is processed and context is canceled
	time.Sleep(100 * time.Millisecond)

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
		t.Skip("skipping on " + windows)
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var mu sync.Mutex
	var exitCode int
	exitCalled := make(chan struct{})
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
		close(exitCalled)
	}

	origTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = origTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	f := &Flow{}
	taskCanFinish := make(chan struct{})
	f.Define(Task{
		Name: task,
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	go f.Main([]string{task})

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)

	// Wait a bit to ensure the first signal is processed
	time.Sleep(100 * time.Millisecond)

	_ = p.Signal(os.Interrupt)

	<-exitCalled

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
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: task,
	})

	f.Main([]string{task})

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}
