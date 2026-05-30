package goyek

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"
)

type safeBuffer struct {
	mu sync.Mutex
	strings.Builder
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Builder.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Builder.String()
}

func TestFlow_Main_signal_graceful(t *testing.T) {
	flow := &Flow{}
	out := &safeBuffer{}
	flow.SetOutput(out)

	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
			<-taskCanFinish
		},
	})

	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		sigChan = c
	}
	defer func() { signalNotify = signal.Notify }()

	signalStop = func(_ chan<- os.Signal) {}
	defer func() { signalStop = signal.Stop }()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}
	defer func() { osExit = os.Exit }()

	done := make(chan struct{})
	handlerReady := make(chan struct{})
	trapSignalsHook = func() { close(handlerReady) }
	defer func() { trapSignalsHook = func() {} }()

	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	<-handlerReady
	sigChan <- os.Interrupt

	close(taskCanFinish)
	<-done

	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()

	if gotExitCode != 1 {
		t.Errorf("exit code: got %d, want 1", gotExitCode)
	}
	if !strings.Contains(out.String(), "first interrupt, graceful stop") {
		t.Errorf("output should contain first interrupt message, got: %s", out.String())
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	flow := &Flow{}
	out := &safeBuffer{}
	flow.SetOutput(out)

	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
			select {} // block forever
		},
	})

	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		sigChan = c
	}
	defer func() { signalNotify = signal.Notify }()

	signalStop = func(_ chan<- os.Signal) {}
	defer func() { signalStop = signal.Stop }()

	var mu sync.Mutex
	exitCode := -1
	osExitCalled := make(chan struct{})
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
		if code == 1 {
			select {
			case osExitCalled <- struct{}{}:
			default:
			}
		}
	}
	defer func() { osExit = os.Exit }()

	handlerReady := make(chan struct{})
	trapSignalsHook = func() { close(handlerReady) }
	defer func() { trapSignalsHook = func() {} }()

	secondStageReady := make(chan struct{})
	trapSignalsSecondHook = func() { close(secondStageReady) }
	defer func() { trapSignalsSecondHook = func() {} }()

	go func() {
		flow.Main([]string{"task"})
	}()

	<-handlerReady
	sigChan <- os.Interrupt
	<-secondStageReady
	sigChan <- os.Interrupt

	<-osExitCalled

	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()

	if gotExitCode != 1 {
		t.Errorf("exit code: got %d, want 1", gotExitCode)
	}
	if !strings.Contains(out.String(), "second interrupt, exit") {
		t.Errorf("output should contain second interrupt message, got: %s", out.String())
	}
}

func TestFlow_Main_pass(t *testing.T) {
	flow := &Flow{}
	out := &safeBuffer{}
	flow.SetOutput(out)
	flow.Define(Task{Name: "task"})

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}
	defer func() { osExit = os.Exit }()

	flow.Main([]string{"task"})

	mu.Lock()
	gotExitCode := exitCode
	mu.Unlock()

	if gotExitCode != 0 {
		t.Errorf("exit code: got %d, want 0", gotExitCode)
	}
}

func TestFlow_Main_default_hooks(_ *testing.T) {
	flow := &Flow{}
	flow.SetOutput(os.Stdout) // Just to make sure it doesn't crash
	flow.Define(Task{Name: "task"})

	// We don't want to actually call os.Exit in this test.
	osExit = func(_ int) {}
	defer func() { osExit = os.Exit }()

	flow.Main([]string{"task"})
}
