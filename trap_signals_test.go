package goyek

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type safeBuffer struct {
	mu sync.Mutex
	sb strings.Builder
}

func (b *safeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.String()
}

func TestFlow_Main_signal_graceful(t *testing.T) {
	flowMu.Lock()
	oldExit := osExit
	oldHook := trapSignalsHook
	defer func() {
		flowMu.Lock()
		osExit = oldExit
		trapSignalsHook = oldHook
		flowMu.Unlock()
	}()

	var exitCode int
	var exitCodeMu sync.Mutex
	osExit = func(code int) {
		exitCodeMu.Lock()
		exitCode = code
		exitCodeMu.Unlock()
	}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}
	flowMu.Unlock()

	f := &Flow{}
	out := &safeBuffer{}
	f.SetOutput(out)

	taskCanFinish := make(chan struct{})
	f.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(done)
	}()

	<-hookCalled

	// Send first signal
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(os.Interrupt)

	// Wait for the message
	for i := 0; i < 1000; i++ {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
		time.Sleep(time.Millisecond)
	}

	// Allow task to finish
	close(taskCanFinish)

	<-done

	exitCodeMu.Lock()
	gotCode := exitCode
	exitCodeMu.Unlock()

	if gotCode != exitCodeFail {
		t.Errorf("expected exit code %d, got %d", exitCodeFail, gotCode)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	flowMu.Lock()
	oldExit := osExit
	oldHook := trapSignalsHook
	defer func() {
		flowMu.Lock()
		osExit = oldExit
		trapSignalsHook = oldHook
		flowMu.Unlock()
	}()

	var exitCode int
	var exitCodeMu sync.Mutex
	exitCalled := make(chan struct{})
	var exitOnce sync.Once
	osExit = func(code int) {
		exitCodeMu.Lock()
		exitCode = code
		exitCodeMu.Unlock()
		exitOnce.Do(func() {
			close(exitCalled)
		})
	}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}
	flowMu.Unlock()

	f := &Flow{}
	out := &safeBuffer{}
	f.SetOutput(out)

	taskCanFinish := make(chan struct{})
	f.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(done)
	}()

	<-hookCalled

	// Send first signal
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(os.Interrupt)

	// Wait for the message
	for i := 0; i < 1000; i++ {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
		time.Sleep(time.Millisecond)
	}

	// Send second signal
	_ = process.Signal(os.Interrupt)

	<-exitCalled

	// Cleanup task
	close(taskCanFinish)
	<-done

	exitCodeMu.Lock()
	gotCode := exitCode
	exitCodeMu.Unlock()

	if gotCode != exitCodeFail {
		t.Errorf("expected exit code %d, got %d", exitCodeFail, gotCode)
	}
	if !strings.Contains(out.String(), "second interrupt, exit") {
		t.Errorf("expected output to contain 'second interrupt, exit'")
	}
}

func TestFlow_Main_pass(t *testing.T) {
	flowMu.Lock()
	oldExit := osExit
	oldHook := trapSignalsHook
	defer func() {
		flowMu.Lock()
		osExit = oldExit
		trapSignalsHook = oldHook
		flowMu.Unlock()
	}()

	var exitCode int
	var exitCodeMu sync.Mutex
	osExit = func(code int) {
		exitCodeMu.Lock()
		exitCode = code
		exitCodeMu.Unlock()
	}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}
	flowMu.Unlock()

	f := &Flow{}
	out := &safeBuffer{}
	f.SetOutput(out)

	f.Define(Task{
		Name: "task",
	})

	done := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(done)
	}()

	<-hookCalled
	<-done

	exitCodeMu.Lock()
	gotCode := exitCode
	exitCodeMu.Unlock()

	if gotCode != exitCodePass {
		t.Errorf("expected exit code %d, got %d", exitCodePass, gotCode)
	}
}
