package goyek

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFlow_Main_signal_graceful(t *testing.T) {
	flow := &Flow{}
	flow.Define(Task{
		Name: "task",
		Action: func(a *A) {
			select {
			case <-a.Context().Done():
			case <-time.After(time.Second):
			}
		},
	})
	sb := &strings.Builder{}
	flow.SetOutput(sb)

	// Mock osExit
	originalOsExit := osExit
	var mu sync.Mutex
	var exitCode int
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}
	defer func() {
		mu.Lock()
		osExit = originalOsExit
		mu.Unlock()
	}()

	// Mock trapSignalsHook
	originalTrapSignalsHook := trapSignalsHook
	sigChan := make(chan os.Signal, 1)
	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			s := <-sigChan
			c <- s
		}()
	}
	defer func() {
		mu.Lock()
		trapSignalsHook = originalTrapSignalsHook
		mu.Unlock()
	}()

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	// Send signal
	sigChan <- os.Interrupt

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Main to return")
	}

	mu.Lock()
	gotCode := exitCode
	mu.Unlock()
	if gotCode != 1 {
		t.Errorf("expected exit code 1, got %d", gotCode)
	}

	if !strings.Contains(sb.String(), "first interrupt, graceful stop") {
		t.Errorf("expected output to contain 'first interrupt, graceful stop', got %q", sb.String())
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	flow := &Flow{}
	sb := &strings.Builder{}
	flow.SetOutput(sb)

	// Mock osExit
	originalOsExit := osExit
	var mu sync.Mutex
	var exitCode int
	var exitCalled bool
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		exitCalled = true
		mu.Unlock()
	}
	defer func() {
		mu.Lock()
		osExit = originalOsExit
		mu.Unlock()
	}()

	// Mock trapSignalsHook
	originalTrapSignalsHook := trapSignalsHook
	sigChan := make(chan os.Signal, 1)
	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for s := range sigChan {
				c <- s
			}
		}()
	}
	defer func() {
		mu.Lock()
		trapSignalsHook = originalTrapSignalsHook
		mu.Unlock()
	}()

	// Start Main in a way it doesn't return immediately
	flow.Define(Task{
		Name: "long",
		Action: func(a *A) {
			select {
			case <-a.Context().Done():
				// wait a bit to simulate work during cleanup
				time.Sleep(100 * time.Millisecond)
			case <-time.After(time.Second):
			}
		},
	})

	mainDone := make(chan struct{})
	go func() {
		flow.Main([]string{"long"})
		close(mainDone)
	}()

	// Send first signal
	sigChan <- os.Interrupt

	// Wait for first message
	time.Sleep(50 * time.Millisecond)

	// Send second signal
	sigChan <- os.Interrupt

	// Wait for osExit to be called
	var called bool
	for i := 0; i < 20; i++ {
		mu.Lock()
		called = exitCalled
		mu.Unlock()
		if called {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !called {
		t.Fatal("osExit was not called")
	}

	mu.Lock()
	gotCode := exitCode
	mu.Unlock()
	if gotCode != 1 {
		t.Errorf("expected exit code 1, got %d", gotCode)
	}

	if !strings.Contains(sb.String(), "second interrupt, exit") {
		t.Errorf("expected output to contain 'second interrupt, exit', got %q", sb.String())
	}

	// Clean up goroutine
	close(sigChan)
	<-mainDone
}

func TestFlow_main_ctx_err(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{
		Name: "test",
		Action: func(a *A) {
			// do nothing
		},
	})

	t.Run("interrupted", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		code := flow.main(ctx, []string{"test"})
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})

	t.Run("passed", func(t *testing.T) {
		code := flow.main(context.Background(), []string{"test"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
}
