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

	restore := mockOsExit()
	defer restore()

	var mu sync.Mutex
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	f := &Flow{}
	taskRun := false
	f.Define(Task{
		Name: "test",
		Action: func(a *A) {
			taskRun = true
			<-a.Context().Done()
		},
	})

	trapSignalsHook = func() {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}
	defer func() { trapSignalsHook = func() {} }()

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"test"})
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	if !taskRun {
		t.Error("task was not run")
	}
	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}


func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	restore := mockOsExit()
	defer restore()

	var mu sync.Mutex
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	f := DefaultFlow
	oldTasks := f.tasks
	f.tasks = make(map[string]*DefinedTask)
	defer func() { f.tasks = oldTasks }()

	taskRun := false
	Define(Task{
		Name: "test",
		Action: func(a *A) {
			taskRun = true
			<-a.Context().Done()
		},
	})

	trapSignalsHook = func() {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}
	defer func() { trapSignalsHook = func() {} }()

	doneCh := make(chan struct{})
	go func() {
		Main([]string{"test"})
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	if !taskRun {
		t.Error("task was not run")
	}
	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	restore := mockOsExit()
	defer restore()

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
		Name: "test",
		Action: func(_ *A) {
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"test"})
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodePass)
	}
}

func TestFlow_Main_signal_hard_timeout(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping on windows")
	}

	restore := mockOsExit()
	defer restore()

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
		Name: "test",
		Action: func(a *A) {
			<-a.Context().Done()
			time.Sleep(time.Hour)
		},
	})

	trapSignalsHook = func() {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}
	trapSignalsSecondHook = func() {
		go func() {
			time.Sleep(100 * time.Millisecond)
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(os.Interrupt)
		}()
	}
	defer func() {
		trapSignalsHook = func() {}
		trapSignalsSecondHook = func() {}
	}()

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"test"})
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(time.Second):
		// This is expected as hard exit will call osExit which might not return
	}

	mu.Lock()
	defer mu.Unlock()
	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func mockOsExit() func() {
	orig := osExit
	return func() { osExit = orig }
}
