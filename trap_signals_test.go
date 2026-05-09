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

	restore := osExit
	defer func() { osExit = restore }()

	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	trapSignalsHook = func() {
		defer wg.Done()
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			return
		}
		_ = p.Signal(os.Interrupt)
	}
	defer func() { trapSignalsHook = nil }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "test",
		Action: func(a *A) {
			wg.Wait()
			<-a.Context().Done()
		},
	})

	f.Main([]string{"test"})

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodeFail {
		t.Errorf("got exit code %d, want %d", got, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	restore := osExit
	defer func() { osExit = restore }()

	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	trapSignalsHook = func() {
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			return
		}
		_ = p.Signal(os.Interrupt)
		// Wait a bit to ensure the first signal is processed and we enter the second select.
		time.Sleep(100 * time.Millisecond)
		_ = p.Signal(os.Interrupt)
	}
	defer func() { trapSignalsHook = nil }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "test",
		Action: func(a *A) {
			<-a.Context().Done()
			// Wait for the second signal to trigger osExit
			time.Sleep(200 * time.Millisecond)
		},
	})

	f.Main([]string{"test"})

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodeFail {
		t.Errorf("got exit code %d, want %d", got, exitCodeFail)
	}
}

func TestFlow_Main_signal_none(t *testing.T) {
	restore := osExit
	defer func() { osExit = restore }()

	var exitCode int
	var mu sync.Mutex
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
	got := exitCode
	mu.Unlock()
	if got != exitCodePass {
		t.Errorf("got exit code %d, want %d", got, exitCodePass)
	}
}

func TestMain_top_level(t *testing.T) {
	restore := osExit
	defer func() { osExit = restore }()

	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	oldOut := DefaultFlow.Output()
	DefaultFlow.SetOutput(io.Discard)
	defer DefaultFlow.SetOutput(oldOut)

	Main([]string{"non-existing-task"})

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodeInvalid {
		t.Errorf("got exit code %d, want %d", got, exitCodeInvalid)
	}
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "task"}
	got := err.Error()
	want := "task failed: task"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
