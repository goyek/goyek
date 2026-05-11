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
	defer func() { osExit = origOsExit }()

	var mu sync.Mutex
	exitCode := -1
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
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "test",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})

	done := make(chan struct{})
	go func() {
		f.Main([]string{"test"})
		close(done)
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-done

	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()
	if gotExitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", gotExitCode, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var mu sync.Mutex
	exitCode := -1
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
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "test",
		Action: func(_ *A) {
			select {} // block forever
		},
	})

	go f.Main([]string{"test"})

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// Wait a bit for the first signal to be processed and context canceled
	// (though in this case it doesn't matter much as we'll send another signal)

	// Send second signal
	// We might need to retry because the handler might not be in the second select yet
	for {
		if err := p.Signal(os.Interrupt); err != nil {
			t.Fatal(err)
		}
		runtime.Gosched()
		mu.Lock()
		code := exitCode
		mu.Unlock()
		if code == exitCodeFail {
			break
		}
	}
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var mu sync.Mutex
	exitCode := -1
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

	// Use DefaultFlow
	task := Define(Task{
		Name: "top-level-test",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})
	defer Undefine(task)

	done := make(chan struct{})
	go func() {
		Main([]string{"top-level-test"})
		close(done)
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-done

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

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "test"})

	f.Main([]string{"test"})

	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()
	if gotExitCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", gotExitCode, exitCodePass)
	}
}
