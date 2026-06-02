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

type signalTestSetup struct {
	mu                      sync.Mutex
	exitCode                int
	exitCalled              bool
	sigChan                 chan os.Signal
	sb                      *strings.Builder
	originalOsExit          func(int)
	originalTrapSignalsHook func(chan<- os.Signal)
}

func setupSignalTest(t *testing.T, flow *Flow) *signalTestSetup {
	t.Helper()
	s := &signalTestSetup{
		sigChan:                 make(chan os.Signal, 1),
		sb:                      &strings.Builder{},
		originalOsExit:          osExit,
		originalTrapSignalsHook: trapSignalsHook,
	}
	flow.SetOutput(s.sb)

	osExit = func(code int) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.exitCode = code
		s.exitCalled = true
	}

	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for sig := range s.sigChan {
				c <- sig
			}
		}()
	}

	return s
}

func (s *signalTestSetup) teardown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	osExit = s.originalOsExit
	trapSignalsHook = s.originalTrapSignalsHook
	close(s.sigChan)
}

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

	s := setupSignalTest(t, flow)
	defer s.teardown()

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	// Send signal
	s.sigChan <- os.Interrupt

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Main to return")
	}

	s.mu.Lock()
	gotCode := s.exitCode
	s.mu.Unlock()
	if gotCode != 1 {
		t.Errorf("expected exit code 1, got %d", gotCode)
	}

	if !strings.Contains(s.sb.String(), "first interrupt, graceful stop") {
		t.Errorf("expected output to contain 'first interrupt, graceful stop', got %q", s.sb.String())
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	flow := &Flow{}
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

	s := setupSignalTest(t, flow)
	defer s.teardown()

	mainDone := make(chan struct{})
	go func() {
		flow.Main([]string{"long"})
		close(mainDone)
	}()

	// Send first signal
	s.sigChan <- os.Interrupt

	// Wait for first message
	time.Sleep(50 * time.Millisecond)

	// Send second signal
	s.sigChan <- os.Interrupt

	// Wait for osExit to be called
	var called bool
	for i := 0; i < 20; i++ {
		s.mu.Lock()
		called = s.exitCalled
		s.mu.Unlock()
		if called {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !called {
		t.Fatal("osExit was not called")
	}

	s.mu.Lock()
	gotCode := s.exitCode
	s.mu.Unlock()
	if gotCode != 1 {
		t.Errorf("expected exit code 1, got %d", gotCode)
	}

	if !strings.Contains(s.sb.String(), "second interrupt, exit") {
		t.Errorf("expected output to contain 'second interrupt, exit', got %q", s.sb.String())
	}

	<-mainDone
}

func TestFlow_main_ctx_err(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{
		Name: "test",
		Action: func(_ *A) {
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
