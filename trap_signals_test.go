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

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sb.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sb.String()
}

func TestFlow_Main_signal_graceful(t *testing.T) {
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var exitCode int
	var exitMu sync.Mutex
	osExit = func(code int) {
		exitMu.Lock()
		exitCode = code
		exitMu.Unlock()
	}

	var sigChan chan<- os.Signal
	var sigMu sync.Mutex
	trapSignalsHook = func(c chan<- os.Signal) {
		sigMu.Lock()
		sigChan = c
		sigMu.Unlock()
	}

	out := &safeBuffer{}
	flow := &Flow{}
	flow.SetOutput(out)

	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	// wait until sigChan is set
	var sc chan<- os.Signal
	for {
		sigMu.Lock()
		sc = sigChan
		sigMu.Unlock()
		if sc != nil {
			break
		}
		time.Sleep(time.Millisecond)
	}
	sc <- os.Interrupt

	// check if "graceful stop" was printed
	for !strings.Contains(out.String(), "first interrupt, graceful stop") {
		time.Sleep(time.Millisecond)
	}

	close(taskCanFinish)
	<-done

	exitMu.Lock()
	if exitCode != 0 {
		t.Errorf("got exit code %d, want 0", exitCode)
	}
	exitMu.Unlock()
}

func TestFlow_Main_signal_hard(t *testing.T) {
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var exitCode int
	var exitMu sync.Mutex
	osExit = func(code int) {
		exitMu.Lock()
		exitCode = code
		exitMu.Unlock()
	}

	var sigChan chan<- os.Signal
	var sigMu sync.Mutex
	trapSignalsHook = func(c chan<- os.Signal) {
		sigMu.Lock()
		sigChan = c
		sigMu.Unlock()
	}

	out := &safeBuffer{}
	flow := &Flow{}
	flow.SetOutput(out)
	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
			select {} // block forever
		},
	})

	go flow.Main([]string{"task"})

	// wait until sigChan is set and send first interrupt
	var sc chan<- os.Signal
	for {
		sigMu.Lock()
		sc = sigChan
		sigMu.Unlock()
		if sc != nil {
			break
		}
		time.Sleep(time.Millisecond)
	}
	sc <- os.Interrupt
	for !strings.Contains(out.String(), "first interrupt, graceful stop") {
		time.Sleep(time.Millisecond)
	}

	// send second interrupt
	sc <- os.Interrupt

	for !strings.Contains(out.String(), "second interrupt, exit") {
		time.Sleep(time.Millisecond)
	}

	// Wait a bit for osExit to be called
	for {
		exitMu.Lock()
		code := exitCode
		exitMu.Unlock()
		if code != 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	exitMu.Lock()
	if exitCode != 1 {
		t.Errorf("got exit code %d, want 1", exitCode)
	}
	exitMu.Unlock()
}

func TestMain_signal_graceful(t *testing.T) {
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	origDefaultFlow := DefaultFlow
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		DefaultFlow = origDefaultFlow
	}()

	var exitCode int
	var exitMu sync.Mutex
	osExit = func(code int) {
		exitMu.Lock()
		exitCode = code
		exitMu.Unlock()
	}

	var sigChan chan<- os.Signal
	var sigMu sync.Mutex
	trapSignalsHook = func(c chan<- os.Signal) {
		sigMu.Lock()
		sigChan = c
		sigMu.Unlock()
	}

	out := &safeBuffer{}
	DefaultFlow = &Flow{}
	DefaultFlow.SetOutput(out)

	taskCanFinish := make(chan struct{})
	Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(done)
	}()

	// wait until sigChan is set
	var sc chan<- os.Signal
	for {
		sigMu.Lock()
		sc = sigChan
		sigMu.Unlock()
		if sc != nil {
			break
		}
		time.Sleep(time.Millisecond)
	}
	sc <- os.Interrupt

	// check if "graceful stop" was printed
	for !strings.Contains(out.String(), "first interrupt, graceful stop") {
		time.Sleep(time.Millisecond)
	}

	close(taskCanFinish)
	<-done

	exitMu.Lock()
	if exitCode != 0 {
		t.Errorf("got exit code %d, want 0", exitCode)
	}
	exitMu.Unlock()
}
