package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
)

const (
	windows = "windows"
	task    = "task"
)

func TestFlow_Main_signal_graceful(t *testing.T) {
	origOsExit := osExit
	origSignalNotify := signalNotify
	origSignalStop := signalStop
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		signalNotify = origSignalNotify
		signalStop = origSignalStop
		trapSignalsHook = origTrapSignalsHook
	}()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	var muSig sync.Mutex
	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		muSig.Lock()
		defer muSig.Unlock()
		sigChan = c
	}
	signalStop = func(_ chan<- os.Signal) {}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: task,
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{task})
		close(done)
	}()

	// Wait for the signal handler to be initialized
	var sc chan<- os.Signal
	for {
		muSig.Lock()
		sc = sigChan
		muSig.Unlock()
		if sc != nil {
			break
		}
		runtime.Gosched()
	}

	sc <- os.Interrupt
	<-hookCalled

	close(taskCanFinish)
	<-done

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestFlow_Main_signal_hard(_ *testing.T) {
	origOsExit := osExit
	origSignalNotify := signalNotify
	origSignalStop := signalStop
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		signalNotify = origSignalNotify
		signalStop = origSignalStop
		trapSignalsHook = origTrapSignalsHook
	}()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	var muSig sync.Mutex
	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		muSig.Lock()
		defer muSig.Unlock()
		sigChan = c
	}
	signalStop = func(_ chan<- os.Signal) {}

	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{
		Name: task,
		Action: func(_ *A) {
			select {} // block forever
		},
	})

	go flow.Main([]string{task})

	// Wait for the signal handler to be initialized
	var sc chan<- os.Signal
	for {
		muSig.Lock()
		sc = sigChan
		muSig.Unlock()
		if sc != nil {
			break
		}
		runtime.Gosched()
	}

	sc <- os.Interrupt
	<-hookCalled
	sc <- os.Interrupt

	// We can't wait on Main's done channel here because Main will call osExit,
	// and in our mock it just sets a variable and returns, so Main will continue
	// and then potentially wait on handlerFinished, but the signal handler
	// goroutine might still be running.
	// However, osExit(exitCodeFail) is called before wait on handlerFinished.
	// Actually, osExit is the LAST thing Main does.
	// Wait for exitCode to be set
	for {
		mu.Lock()
		code := exitCode
		mu.Unlock()
		if code == 1 {
			break
		}
		runtime.Gosched()
	}
}

func TestMain_signal_graceful(t *testing.T) {
	origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook := osExit, signalNotify, signalStop, trapSignalsHook
	defer func() {
		osExit, signalNotify, signalStop, trapSignalsHook = origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook
	}()
	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}
	var muSig sync.Mutex
	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		muSig.Lock()
		defer muSig.Unlock()
		sigChan = c
	}
	signalStop = func(_ chan<- os.Signal) {}
	hookCalled := make(chan struct{})
	trapSignalsHook = func() { close(hookCalled) }
	origDefaultFlow := DefaultFlow
	defer func() { DefaultFlow = origDefaultFlow }()
	DefaultFlow = &Flow{}
	DefaultFlow.SetOutput(io.Discard)
	taskCanFinish := make(chan struct{})
	DefaultFlow.Define(Task{
		Name: task,
		Action: func(_ *A) { <-taskCanFinish },
	})

	done := make(chan struct{})
	go func() {
		Main([]string{task})
		close(done)
	}()

	// Wait for the signal handler to be initialized
	var sc chan<- os.Signal
	for {
		muSig.Lock()
		sc = sigChan
		muSig.Unlock()
		if sc != nil {
			break
		}
		runtime.Gosched()
	}

	sc <- os.Interrupt
	<-hookCalled

	close(taskCanFinish)
	<-done

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var mu sync.Mutex
	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: task})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{task})
		close(done)
	}()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}
