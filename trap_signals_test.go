package goyek

import (
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestFlow_Main_signal_graceful(t *testing.T) {
	origSignalNotify := signalNotify
	origSignalStop := signalStop
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		signalNotify = origSignalNotify
		signalStop = origSignalStop
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var sigChan chan<- os.Signal
	var mu sync.Mutex
	signalNotify = func(c chan<- os.Signal, sigs ...os.Signal) {
		mu.Lock()
		sigChan = c
		mu.Unlock()
	}
	signalStop = func(c chan<- os.Signal) {}

	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}

	flow := &Flow{}
	out := &strings.Builder{}
	flow.SetOutput(out)

	taskStarted := make(chan struct{})
	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			close(taskStarted)
			<-taskCanFinish
		},
	})

	trapSignalsHook = func() {
		for {
			mu.Lock()
			c := sigChan
			mu.Unlock()
			if c != nil {
				c <- os.Interrupt
				return
			}
		}
	}

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	<-taskStarted
	// sigChan should be set by signalNotify called within Main's goroutine
	// trapSignalsHook is called from within Main's goroutine

	// Ensure that f.main finished after cancellation
	// Wait a bit to ensure the signal is processed
	for {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
	}
	close(taskCanFinish)
	<-done

	if atomic.LoadInt32(&exitCode) != 1 {
		t.Errorf("exit code should be 1, got: %d", atomic.LoadInt32(&exitCode))
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	origSignalNotify := signalNotify
	origSignalStop := signalStop
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	origTrapSignalsSecondHook := trapSignalsSecondHook
	defer func() {
		signalNotify = origSignalNotify
		signalStop = origSignalStop
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		trapSignalsSecondHook = origTrapSignalsSecondHook
	}()

	var sigChan chan<- os.Signal
	var mu sync.Mutex
	signalNotify = func(c chan<- os.Signal, sigs ...os.Signal) {
		mu.Lock()
		sigChan = c
		mu.Unlock()
	}
	signalStop = func(c chan<- os.Signal) {}

	var exitCode int32 = -1
	osExit = func(code int) {
		mu.Lock()
		if exitCode == -1 {
			atomic.StoreInt32(&exitCode, int32(code))
		}
		mu.Unlock()
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)

	taskStarted := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			close(taskStarted)
			select {} // block forever
		},
	})

	trapSignalsHook = func() {
		for {
			mu.Lock()
			c := sigChan
			mu.Unlock()
			if c != nil {
				c <- os.Interrupt
				return
			}
		}
	}
	trapSignalsSecondHook = func() {
		for {
			mu.Lock()
			c := sigChan
			mu.Unlock()
			if c != nil {
				c <- os.Interrupt
				return
			}
		}
	}

	go flow.Main([]string{"task"})

	<-taskStarted
	// Wait for osExit to be called
	for {
		mu.Lock()
		code := atomic.LoadInt32(&exitCode)
		mu.Unlock()
		if code != -1 {
			break
		}
	}

	if atomic.LoadInt32(&exitCode) != 1 {
		t.Errorf("exit code should be 1, got: %d", atomic.LoadInt32(&exitCode))
	}
}

func TestMain_signal_graceful(t *testing.T) {
	origSignalNotify := signalNotify
	origSignalStop := signalStop
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		signalNotify = origSignalNotify
		signalStop = origSignalStop
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var sigChan chan<- os.Signal
	var mu sync.Mutex
	signalNotify = func(c chan<- os.Signal, sigs ...os.Signal) {
		mu.Lock()
		sigChan = c
		mu.Unlock()
	}
	signalStop = func(c chan<- os.Signal) {}

	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}

	DefaultFlow = &Flow{} // reset DefaultFlow
	out := &strings.Builder{}
	SetOutput(out)

	taskStarted := make(chan struct{})
	taskCanFinish := make(chan struct{})
	Define(Task{
		Name: "task",
		Action: func(a *A) {
			close(taskStarted)
			<-taskCanFinish
		},
	})

	trapSignalsHook = func() {
		for {
			mu.Lock()
			c := sigChan
			mu.Unlock()
			if c != nil {
				c <- os.Interrupt
				return
			}
		}
	}

	done := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(done)
	}()

	<-taskStarted
	// sigChan should be set by signalNotify called within Main's goroutine
	// trapSignalsHook is called from within Main's goroutine

	// Wait a bit to ensure the signal is processed
	for {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
	}

	close(taskCanFinish)
	<-done

	if atomic.LoadInt32(&exitCode) != 1 {
		t.Errorf("exit code should be 1, got: %d", atomic.LoadInt32(&exitCode))
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int32 = -1
	osExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	flow.Main([]string{"task"})

	if atomic.LoadInt32(&exitCode) != 0 {
		t.Errorf("exit code should be 0, got: %d", atomic.LoadInt32(&exitCode))
	}
}
