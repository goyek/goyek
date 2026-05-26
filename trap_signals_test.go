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
	origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook, origTrapSignalsSecondHook := osExit, signalNotify, signalStop, trapSignalsHook, trapSignalsSecondHook
	defer func() {
		osExit, signalNotify, signalStop, trapSignalsHook, trapSignalsSecondHook = origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook, origTrapSignalsSecondHook
	}()

	var mu sync.Mutex
	osExit = func(_ int) {
		mu.Lock()
		defer mu.Unlock()
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

	secondHookCalled := make(chan struct{})
	trapSignalsSecondHook = func() { close(secondHookCalled) }

	sc <- os.Interrupt
	<-secondHookCalled

	close(taskCanFinish)
	<-done
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
	sc, hookCalled, taskCanFinish, done := setupMainTest()
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

func setupMainTest() (chan<- os.Signal, <-chan struct{}, chan struct{}, <-chan struct{}) {
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
		DefaultFlow = origDefaultFlow
		close(done)
	}()
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
	return sc, hookCalled, taskCanFinish, done
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

func TestFlow_Main_default_hooks(_ *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	osExit = func(code int) {}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: task})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{task})
		close(done)
	}()
	<-done
}
