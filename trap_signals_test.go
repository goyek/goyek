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
	s  string
}

func (b *safeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.s += string(p)
	return len(p), nil
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.s
}

func TestFlow_Main_signal_graceful(t *testing.T) {
	flowMu.Lock()
	defer flowMu.Unlock()

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var exitCode int
	osExit = func(code int) { exitCode = code }

	sigChan := make(chan os.Signal, 1)
	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for sig := range sigChan {
				c <- sig
			}
		}()
	}

	out := &safeBuffer{}
	flow := &Flow{}
	flow.SetOutput(out)

	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	// wait for task to start
	time.Sleep(100 * time.Millisecond)

	// first signal
	sigChan <- os.Interrupt

	// wait for graceful stop message
	waitForOutput(t, out, "first interrupt, graceful stop")

	// allow task to finish
	close(taskCanFinish)

	<-done

	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	flowMu.Lock()
	defer flowMu.Unlock()

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var exitCode int
	var exitWg sync.WaitGroup
	exitWg.Add(1)
	osExit = func(code int) {
		exitCode = code
		exitWg.Done()
	}

	sigChan := make(chan os.Signal, 1)
	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for sig := range sigChan {
				c <- sig
			}
		}()
	}

	out := &safeBuffer{}
	flow := &Flow{}
	flow.SetOutput(out)

	flow.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			select {} // block forever
		},
	})

	go flow.Main([]string{"task"})

	// wait for task to start
	time.Sleep(100 * time.Millisecond)

	// first signal
	sigChan <- os.Interrupt
	waitForOutput(t, out, "first interrupt, graceful stop")

	// second signal
	sigChan <- os.Interrupt
	waitForOutput(t, out, "second interrupt, exit")

	exitWg.Wait()

	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	flowMu.Lock()
	defer flowMu.Unlock()

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	defer func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
	}()

	var exitCode int
	osExit = func(code int) { exitCode = code }

	sigChan := make(chan os.Signal, 1)
	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for sig := range sigChan {
				c <- sig
			}
		}()
	}

	out := &safeBuffer{}
	oldDefaultFlow := DefaultFlow
	defer func() { DefaultFlow = oldDefaultFlow }()
	DefaultFlow = &Flow{} // Reset DefaultFlow
	SetOutput(out)

	taskCanFinish := make(chan struct{})
	Define(Task{
		Name: "task",
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(done)
	}()

	// wait for task to start
	time.Sleep(100 * time.Millisecond)

	// first signal
	sigChan <- os.Interrupt
	waitForOutput(t, out, "first interrupt, graceful stop")

	// allow task to finish
	close(taskCanFinish)

	<-done

	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	flowMu.Lock()
	defer flowMu.Unlock()

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	osExit = func(code int) { exitCode = code }

	flow := &Flow{}
	flow.Define(Task{Name: "task"})

	flow.Main([]string{"task"})

	if exitCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodePass)
	}
}

func waitForOutput(t *testing.T, out *safeBuffer, substr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), substr) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("timeout waiting for output: %q, got: %q", substr, out.String())
}
