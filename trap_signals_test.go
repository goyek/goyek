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

type signalTestSetup struct {
	sigChan  chan os.Signal
	exitCode *int
	exitWg   *sync.WaitGroup
	out      *safeBuffer
}

func setupSignalTest(t *testing.T) *signalTestSetup {
	t.Helper()
	flowMu.Lock()

	origOsExit := osExit
	origTrapSignalsHook := trapSignalsHook
	origDefaultFlow := DefaultFlow

	s := &signalTestSetup{
		sigChan:  make(chan os.Signal, 1),
		exitCode: new(int),
		exitWg:   &sync.WaitGroup{},
		out:      &safeBuffer{},
	}
	s.exitWg.Add(1)

	var once sync.Once
	osExit = func(code int) {
		once.Do(func() {
			*s.exitCode = code
			s.exitWg.Done()
		})
	}

	trapSignalsHook = func(c chan<- os.Signal) {
		go func() {
			for sig := range s.sigChan {
				c <- sig
			}
		}()
	}

	t.Cleanup(func() {
		osExit = origOsExit
		trapSignalsHook = origTrapSignalsHook
		DefaultFlow = origDefaultFlow
		close(s.sigChan)
		flowMu.Unlock()
	})

	return s
}

func TestFlow_Main_signal_graceful(t *testing.T) {
	s := setupSignalTest(t)
	flow := &Flow{}
	flow.SetOutput(s.out)

	taskStarted := make(chan struct{})
	taskCanFinish := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			close(taskStarted)
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		flow.Main([]string{"task"})
		close(done)
	}()

	waitForChan(t, taskStarted, "task to start")
	s.sigChan <- os.Interrupt
	waitForOutput(t, s.out, "first interrupt, graceful stop")

	close(taskCanFinish)
	waitForChan(t, done, "Main to finish")

	if *s.exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", *s.exitCode, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	s := setupSignalTest(t)
	flow := &Flow{}
	flow.SetOutput(s.out)

	taskStarted := make(chan struct{})
	flow.Define(Task{
		Name: "task",
		Action: func(_ *A) {
			close(taskStarted)
			select {} // block forever
		},
	})

	go flow.Main([]string{"task"})

	waitForChan(t, taskStarted, "task to start")
	s.sigChan <- os.Interrupt
	waitForOutput(t, s.out, "first interrupt, graceful stop")

	s.sigChan <- os.Interrupt
	waitForOutput(t, s.out, "second interrupt, exit")

	s.exitWg.Wait()

	if *s.exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", *s.exitCode, exitCodeFail)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	s := setupSignalTest(t)
	DefaultFlow = &Flow{}
	SetOutput(s.out)

	taskStarted := make(chan struct{})
	taskCanFinish := make(chan struct{})
	Define(Task{
		Name: "task",
		Action: func(_ *A) {
			close(taskStarted)
			<-taskCanFinish
		},
	})

	done := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(done)
	}()

	waitForChan(t, taskStarted, "task to start")
	s.sigChan <- os.Interrupt
	waitForOutput(t, s.out, "first interrupt, graceful stop")

	close(taskCanFinish)
	waitForChan(t, done, "Main to finish")

	if *s.exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", *s.exitCode, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	flowMu.Lock()
	defer flowMu.Unlock()
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	var wg sync.WaitGroup
	wg.Add(1)
	osExit = func(code int) {
		exitCode = code
		wg.Done()
	}

	flow := &Flow{}
	flow.Define(Task{Name: "task"})

	flow.Main([]string{"task"})
	wg.Wait()

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

func waitForChan(t *testing.T, ch <-chan struct{}, desc string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s", desc)
	}
}
