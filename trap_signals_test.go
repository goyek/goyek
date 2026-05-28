package goyek

import (
	"os"
	"sync"
	"testing"
)

func TestFlow_Main_signal_graceful(t *testing.T) {
	// Backup and restore
	oldSignalNotify := signalNotify
	oldSignalStop := signalStop
	oldOsExit := osExit
	defer func() {
		signalNotify = oldSignalNotify
		signalStop = oldSignalStop
		osExit = oldOsExit
	}()

	var mu sync.Mutex
	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		mu.Lock()
		sigChan = c
		mu.Unlock()
	}
	signalStop = func(_ chan<- os.Signal) {}

	osExit = func(_ int) {
		// do nothing
	}

	f := &Flow{}
	taskRan := false
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			taskRan = true
			// Wait for signal to be processed
			<-a.Context().Done()
		},
	})

	done := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(done)
	}()

	// Wait for sigChan to be set
	var sc chan<- os.Signal
	for {
		mu.Lock()
		sc = sigChan
		mu.Unlock()
		if sc != nil {
			break
		}
	}

	// Send first signal
	sc <- os.Interrupt

	<-done

	if !taskRan {
		t.Error("task did not run")
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	// Backup and restore
	oldSignalNotify := signalNotify
	oldSignalStop := signalStop
	oldOsExit := osExit
	defer func() {
		signalNotify = oldSignalNotify
		signalStop = oldSignalStop
		osExit = oldOsExit
	}()

	var mu sync.Mutex
	var sigChan chan<- os.Signal
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		mu.Lock()
		sigChan = c
		mu.Unlock()
	}
	signalStop = func(_ chan<- os.Signal) {}

	exitCode := -1
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}

	f := &Flow{}
	f.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			select {} // block forever
		},
	})

	go f.Main([]string{"task"})

	// Wait for sigChan to be set
	var sc chan<- os.Signal
	for {
		mu.Lock()
		sc = sigChan
		mu.Unlock()
		if sc != nil {
			break
		}
	}

	// Send first signal
	sc <- os.Interrupt
	// Send second signal
	sc <- os.Interrupt

	// Wait for osExit to be called
	for {
		mu.Lock()
		code := exitCode
		mu.Unlock()
		if code != -1 {
			break
		}
	}

	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}
