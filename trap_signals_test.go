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
		t.Skip("skipping signal test on windows")
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

	origTrapSignalsHook := trapSignalsHook
	ready := make(chan struct{})
	trapSignalsHook = func() {
		close(ready)
	}
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(doneCh)
	}()

	<-ready
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
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
		defer mu.Unlock()
		exitCode = code
	}

	origTrapSignalsHook := trapSignalsHook
	ready := make(chan struct{})
	trapSignalsHook = func() {
		close(ready)
	}
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	task := f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})
	f.SetDefault(task)

	oldDefaultFlow := DefaultFlow
	DefaultFlow = f
	defer func() { DefaultFlow = oldDefaultFlow }()

	doneCh := make(chan struct{})
	go func() {
		Main(nil)
		close(doneCh)
	}()

	<-ready
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
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
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task"})

	f.Main([]string{"task"})

	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodePass)
	}
}
