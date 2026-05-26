package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
)

const task = "task"

func setupTest(flow *Flow, run func([]string, ...Option)) (chan<- os.Signal, <-chan struct{}, chan struct{}, <-chan struct{}) {
	var muSig sync.Mutex
	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) { muSig.Lock(); defer muSig.Unlock(); sigChan = c }
	signalStop = func(_ chan<- os.Signal) {}
	hookCalled := make(chan struct{})
	trapSignalsHook = func() { close(hookCalled) }
	flow.SetOutput(io.Discard)
	taskCanFinish := make(chan struct{})
	flow.Define(Task{Name: task, Action: func(_ *A) { <-taskCanFinish }})
	done := make(chan struct{})
	go func() { run([]string{task}); close(done) }()
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

func TestFlow_Main_signal_graceful(t *testing.T) {
	origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook := osExit, signalNotify, signalStop, trapSignalsHook
	defer func() { osExit, signalNotify, signalStop, trapSignalsHook = origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook }()
	var mu sync.Mutex
	exitCode := -1
	osExit = func(c int) { mu.Lock(); defer mu.Unlock(); exitCode = c }
	flow := &Flow{}
	sc, hookCalled, taskCanFinish, done := setupTest(flow, flow.Main)
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
	defer func() { osExit, signalNotify, signalStop, trapSignalsHook, trapSignalsSecondHook = origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook, origTrapSignalsSecondHook }()
	osExit = func(_ int) {}
	flow := &Flow{}
	sc, hookCalled, taskCanFinish, done := setupTest(flow, flow.Main)
	sc <- os.Interrupt
	<-hookCalled
	secondHookCalled := make(chan struct{})
	trapSignalsSecondHook = func() { close(secondHookCalled) }
	sc <- os.Interrupt
	<-secondHookCalled
	sc <- os.Interrupt
	close(taskCanFinish)
	<-done
}

func TestMain_signal_graceful(t *testing.T) {
	origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook := osExit, signalNotify, signalStop, trapSignalsHook
	defer func() { osExit, signalNotify, signalStop, trapSignalsHook = origOsExit, origSignalNotify, origSignalStop, origTrapSignalsHook }()
	var mu sync.Mutex
	exitCode := -1
	osExit = func(c int) { mu.Lock(); defer mu.Unlock(); exitCode = c }
	origDefaultFlow := DefaultFlow
	DefaultFlow = &Flow{}
	defer func() { DefaultFlow = origDefaultFlow }()
	sc, hookCalled, taskCanFinish, done := setupTest(DefaultFlow, Main)
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

func TestFlow_Main_pass(_ *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	osExit = func(_ int) {}
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

func TestFlow_Main_default_hooks(_ *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()
	osExit = func(_ int) {}
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
