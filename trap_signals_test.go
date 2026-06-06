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
	flow := &Flow{}
	out := &safeBuffer{}
	flow.SetOutput(out)

	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done() // Wait for interrupt
			<-taskCanFinish      // Wait for signal from test to finish
		},
	})

	flowMu.Lock()
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	var exitCode int
	osExit = func(code int) {
		flowMu.Lock()
		exitCode = code
		flowMu.Unlock()
	}
	sigChan := make(chan os.Signal)
	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for sig := range sigChan {
				c <- sig
			}
		}()
	}
	flowMu.Unlock()

	defer func() {
		flowMu.Lock()
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		flowMu.Unlock()
	}()

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	// Send first signal
	sigChan <- os.Interrupt

	// Check if output contains "graceful stop"
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(out.String(), "first interrupt, graceful stop") {
		t.Errorf("expected graceful stop message, got: %q", out.String())
	}

	// Allow task to finish
	close(taskCanFinish)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Main did not finish in time")
	}

	flowMu.Lock()
	gotCode := exitCode
	flowMu.Unlock()
	if gotCode != 1 {
		t.Errorf("expected exit code 1, got %d", gotCode)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	flow := &Flow{}
	out := &safeBuffer{}
	flow.SetOutput(out)

	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done() // Wait for interrupt
			select {}            // Block forever
		},
	})

	flowMu.Lock()
	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	exitCalled := make(chan int, 1)
	osExit = func(code int) {
		exitCalled <- code
		// In a real app, os.Exit stops everything.
		// In test, we can't really stop the goroutine easily without panic or runtime.Goexit
	}
	sigChan := make(chan os.Signal)
	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for sig := range sigChan {
				c <- sig
			}
		}()
	}
	flowMu.Unlock()

	defer func() {
		flowMu.Lock()
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		flowMu.Unlock()
	}()

	go func() {
		flow.Main([]string{"task"})
	}()

	// Send first signal
	sigChan <- os.Interrupt

	// Wait for graceful stop message
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), "first interrupt, graceful stop") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Send second signal
	sigChan <- os.Interrupt

	select {
	case code := <-exitCalled:
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	case <-time.After(time.Second):
		t.Fatal("os.Exit was not called")
	}

	if !strings.Contains(out.String(), "second interrupt, exit") {
		t.Errorf("expected hard exit message, got: %q", out.String())
	}
}
