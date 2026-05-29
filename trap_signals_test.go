package goyek

import (
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
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

func setupTest() (restore func(), sigChan chan chan<- os.Signal, exitCode *int32, mu *sync.Mutex) {
	origSignalNotify := signalNotify
	origSignalStop := signalStop
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	origTrapSignalsSecondHook := trapSignalsSecondHook
	origDefaultFlow := DefaultFlow

	restore = func() {
		signalNotify = origSignalNotify
		signalStop = origSignalStop
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		trapSignalsSecondHook = origTrapSignalsSecondHook
		DefaultFlow = origDefaultFlow
	}

	sigChan = make(chan chan<- os.Signal, 1)
	mu = &sync.Mutex{}
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		sigChan <- c
	}
	signalStop = func(_ chan<- os.Signal) {}

	var code int32 = -1
	exitCode = &code
	osExit = func(c int) {
		mu.Lock()
		if atomic.LoadInt32(exitCode) == -1 {
			atomic.StoreInt32(exitCode, int32(c)) //nolint:gosec // G115: exit code is a small integer
		}
		mu.Unlock()
	}

	return restore, sigChan, exitCode, mu
}

func TestFlow_Main_signal_graceful(t *testing.T) {
	restore, sigChan, exitCode, _ := setupTest()
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
	for {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
	}
	close(taskCanFinish)
	<-done

	if atomic.LoadInt32(exitCode) != 1 {
		t.Errorf("exit code should be 1, got: %d", atomic.LoadInt32(exitCode))
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	restore, sigChan, exitCode, mu := setupTest()
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
		c <- os.Interrupt
	}

	go flow.Main([]string{"task"})

	<-taskStarted
	for {
		mu.Lock()
		code := atomic.LoadInt32(exitCode)
		mu.Unlock()
		if code != -1 {
			break
		}
	}

	if atomic.LoadInt32(exitCode) != 1 {
		t.Errorf("exit code should be 1, got: %d", atomic.LoadInt32(exitCode))
	}
}

func TestMain_signal_graceful(t *testing.T) {
	restore, sigChan, exitCode, _ := setupTest()
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
	for {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
	}
	close(taskCanFinish)
	<-done

	if atomic.LoadInt32(exitCode) != 1 {
		t.Errorf("exit code should be 1, got: %d", atomic.LoadInt32(exitCode))
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code)) //nolint:gosec // G115: exit code is a small integer
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	flow.Main([]string{"task"})

	if atomic.LoadInt32(&exitCode) != 0 {
		t.Errorf("exit code should be 0, got: %d", atomic.LoadInt32(&exitCode))
	}
}
