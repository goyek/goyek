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

	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	osExit = func(code int) { exitCode = code }

	restoreTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = restoreTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() { close(hookCalled) }

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			select {
			case <-a.Context().Done():
			case <-time.After(time.Second):
				a.Error("timeout")
			}
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(doneCh)
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var mu sync.Mutex
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	restoreTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = restoreTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() { close(hookCalled) }

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			select {
			case <-a.Context().Done():
				// hang to allow second signal
				select {}
			case <-time.After(time.Second):
				a.Error("timeout")
			}
		},
	})

	go f.Main([]string{"task"})

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// Wait for the first signal to be processed and then send the second one repeatedly
	var got int
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if err := p.Signal(os.Interrupt); err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		got = exitCode
		mu.Unlock()
		if got == 1 {
			break
		}
		runtime.Gosched()
	}

	if got != 1 {
		t.Errorf("expected exit code 1, got %d", got)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	osExit = func(code int) { exitCode = code }

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task"})

	f.Main([]string{"task"})

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	osExit = func(code int) { exitCode = code }

	restoreTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = restoreTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() { close(hookCalled) }

	DefaultFlow = &Flow{}
	Define(Task{
		Name: "task",
		Action: func(a *A) {
			select {
			case <-a.Context().Done():
			case <-time.After(time.Second):
				a.Error("timeout")
			}
		},
	})

	doneCh := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(doneCh)
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}
