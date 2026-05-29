package goyek

import (
	"io"
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

func setupTest() (restore func(), sigChan chan chan<- os.Signal, exitCodeCalled chan int) {
	origSignalNotify := signalNotify
	origSignalStop := signalStop
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	origTrapSignalsSecondHook := trapSignalsSecondHook
	origDefaultFlow := DefaultFlow

	exitCodeCalled = make(chan int, 10)

	restore = func() {
		signalNotify = origSignalNotify
		signalStop = origSignalStop
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		trapSignalsSecondHook = origTrapSignalsSecondHook
		DefaultFlow = origDefaultFlow
	}

	sigChan = make(chan chan<- os.Signal, 1)
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		sigChan <- c
	}
	signalStop = func(_ chan<- os.Signal) {}

	osExit = func(c int) {
		exitCodeCalled <- c
	}

	return restore, sigChan, exitCodeCalled
}

func TestFlow_Main_signal_graceful(t *testing.T) {
	restore, sigChan, exitCodeCalled := setupTest()
	defer restore()

	flow := &Flow{}
	out := &safeBuffer{}
	flow.SetOutput(out)

	taskStarted := make(chan struct{})
	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			close(taskStarted)
			<-taskCanFinish
		},
	})

	trapSignalsHook = func() {
		c := <-sigChan
		c <- os.Interrupt
	}

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	<-taskStarted
	// Wait for the graceful stop message to appear
	start := time.Now()
	for {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
		if time.Since(start) > 5*time.Second {
			t.Fatal("timed out waiting for graceful stop message")
		}
		time.Sleep(10 * time.Millisecond)
	}
	close(taskCanFinish)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Main did not return")
	}

	select {
	case code := <-exitCodeCalled:
		if code != 1 {
			t.Errorf("exit code should be 1, got: %d", code)
		}
	default:
		t.Error("os.Exit was not called")
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	restore, sigChan, exitCodeCalled := setupTest()
	defer restore()

	flow := &Flow{}
	flow.SetOutput(io.Discard)

	taskStarted := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			close(taskStarted)
			select {} // block forever
		},
	})

	var c chan<- os.Signal
	trapSignalsHook = func() {
		c = <-sigChan
		c <- os.Interrupt
	}
	trapSignalsSecondHook = func() {
		if c != nil {
			c <- os.Interrupt
		}
	}

	go flow.Main([]string{"task"})

	<-taskStarted

	select {
	case code := <-exitCodeCalled:
		if code != 1 {
			t.Errorf("exit code should be 1, got: %d", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for os.Exit")
	}
}

func TestMain_signal_graceful(t *testing.T) {
	restore, sigChan, exitCodeCalled := setupTest()
	defer restore()

	DefaultFlow = &Flow{}
	out := &safeBuffer{}
	SetOutput(out)

	taskStarted := make(chan struct{})
	taskCanFinish := make(chan struct{})
	Define(Task{
		Name: "task",
		Action: func(_ *A) {
			close(taskStarted)
			<-taskCanFinish
		},
	})

	trapSignalsHook = func() {
		c := <-sigChan
		c <- os.Interrupt
	}

	done := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(done)
	}()

	<-taskStarted
	start := time.Now()
	for {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
		if time.Since(start) > 5*time.Second {
			t.Fatal("timed out waiting for graceful stop message")
		}
		time.Sleep(10 * time.Millisecond)
	}
	close(taskCanFinish)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Main did not return")
	}

	select {
	case code := <-exitCodeCalled:
		if code != 1 {
			t.Errorf("exit code should be 1, got: %d", code)
		}
	default:
		t.Error("os.Exit was not called")
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	exitCode := make(chan int, 1)
	osExit = func(code int) {
		exitCode <- code
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	flow.Main([]string{"task"})

	select {
	case code := <-exitCode:
		if code != 0 {
			t.Errorf("exit code should be 0, got: %d", code)
		}
	default:
		t.Error("os.Exit was not called")
	}
}
