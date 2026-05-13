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
			<-a.Context().Done()
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
	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != 1 {
		t.Errorf("expected exit code 1, got %d", got)
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

	taskFinished := make(chan struct{})
	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
			// Return after a short delay to allow second signal to be caught
			time.Sleep(100 * time.Millisecond)
			close(taskFinished)
		},
	})

	go f.Main([]string{"task"})

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// Wait a bit for the first signal to be processed
	time.Sleep(20 * time.Millisecond)

	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	// Wait for the hard exit to be recorded
	var got int
	for i := 0; i < 100; i++ {
		mu.Lock()
		got = exitCode
		mu.Unlock()
		if got == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got != 1 {
		t.Errorf("expected exit code 1, got %d", got)
	}

	// Ensure the task finished and Main can return
	<-taskFinished
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

	origDefaultFlow := DefaultFlow
	defer func() { DefaultFlow = origDefaultFlow }()
	DefaultFlow = &Flow{}
	Define(Task{
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

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh
	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != 1 {
		t.Errorf("expected exit code 1, got %d", got)
	}
}

func TestFlow_Main_extra_signals(t *testing.T) {
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
			<-a.Context().Done()
			// Hang to allow extra signals
			time.Sleep(150 * time.Millisecond)
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(doneCh)
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())

	// First signal: graceful stop
	_ = p.Signal(os.Interrupt)
	time.Sleep(20 * time.Millisecond)

	// Second signal: hard exit
	_ = p.Signal(os.Interrupt)
	time.Sleep(20 * time.Millisecond)

	// Extra signals: should be consumed by the loop
	for i := 0; i < 5; i++ {
		_ = p.Signal(os.Interrupt)
		time.Sleep(10 * time.Millisecond)
	}

	<-doneCh
	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != 1 {
		t.Errorf("expected exit code 1, got %d", got)
	}
}
