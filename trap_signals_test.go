package goyek

import (
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	windows = "windows"
	task    = "task"
)

func TestFlow_Main_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	restore := mockOsExit()
	defer restore()

	var (
		mu       sync.Mutex
		exitCode int
	)
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	taskCanFinish := make(chan struct{})
	f := &Flow{}
	f.Define(Task{
		Name: task,
		Action: func(a *A) {
			<-taskCanFinish
		},
	})

	trapSignalsHook = func() {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}
	defer func() { trapSignalsHook = func() {} }()

	done := make(chan struct{})
	go func() {
		f.Main([]string{task})
		close(done)
	}()

	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)

	close(taskCanFinish)
	<-done

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodeFail {
		t.Errorf("exit code: got %d, want %d", got, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	restore := mockOsExit()
	defer restore()

	var (
		mu       sync.Mutex
		exitCode int
	)
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	taskStarted := make(chan struct{})
	f := &Flow{}
	f.Define(Task{
		Name: task,
		Action: func(a *A) {
			close(taskStarted)
			select {
			case <-a.Context().Done():
			case <-time.After(10 * time.Second):
			}
		},
	})

	trapSignalsHook = func() {
		go func() {
			<-taskStarted
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(os.Interrupt)
		}()
	}
	defer func() { trapSignalsHook = func() {} }()

	trapSignalsSecondHook = func() {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}
	defer func() { trapSignalsSecondHook = func() {} }()

	done := make(chan struct{})
	go func() {
		f.Main([]string{task})
		close(done)
	}()

	<-done

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodeFail {
		t.Errorf("exit code: got %d, want %d", got, exitCodeFail)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("sending signals is not supported on Windows")
	}

	restore := mockOsExit()
	defer restore()

	var (
		mu       sync.Mutex
		exitCode int
	)
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	taskCanFinish := make(chan struct{})
	origDefaultFlow := DefaultFlow
	DefaultFlow = &Flow{}
	defer func() { DefaultFlow = origDefaultFlow }()

	Define(Task{
		Name: task,
		Action: func(a *A) {
			<-taskCanFinish
		},
	})

	trapSignalsHook = func() {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}
	defer func() { trapSignalsHook = func() {} }()

	done := make(chan struct{})
	go func() {
		Main([]string{task})
		close(done)
	}()

	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)

	close(taskCanFinish)
	<-done

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodeFail {
		t.Errorf("exit code: got %d, want %d", got, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	restore := mockOsExit()
	defer restore()

	var (
		mu       sync.Mutex
		exitCode int
	)
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	f := &Flow{}
	f.Define(Task{
		Name: task,
		Action: func(a *A) {
		},
	})

	f.Main([]string{task})

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodePass {
		t.Errorf("exit code: got %d, want %d", got, exitCodePass)
	}
}

func TestFlow_Main_default_hooks(t *testing.T) {
	restore := mockOsExit()
	defer restore()

	var (
		mu       sync.Mutex
		exitCode int
	)
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	f := &Flow{}
	f.Define(Task{
		Name: task,
		Action: func(a *A) {
		},
	})

	// Just call with default hooks to ensure coverage
	f.Main([]string{task})

	mu.Lock()
	got := exitCode
	mu.Unlock()
	if got != exitCodePass {
		t.Errorf("exit code: got %d, want %d", got, exitCodePass)
	}
}

func mockOsExit() func() {
	orig := osExit
	return func() {
		osExit = orig
	}
}
