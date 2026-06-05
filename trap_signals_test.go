package goyek

import (
	"io"
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
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	var exitMu sync.Mutex
	osExit = func(code int) {
		exitMu.Lock()
		defer exitMu.Unlock()
		exitCode = code
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

	// Wait for the task to start and be blocked
	time.Sleep(100 * time.Millisecond)

	// Send the first interrupt signal
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(os.Interrupt)

	// Wait for the signal handler to process it
	time.Sleep(100 * time.Millisecond)

	// Allow the task to finish
	close(taskCanFinish)

	<-done

	exitMu.Lock()
	gotCode := exitCode
	exitMu.Unlock()

	if gotCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", gotCode, exitCodeFail)
	}

	gotOutput := out.String()
	if !strings.Contains(gotOutput, "first interrupt, graceful stop") {
		t.Errorf("output should contain graceful stop message, got: %q", gotOutput)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	var exitMu sync.Mutex
	osExit = func(code int) {
		exitMu.Lock()
		defer exitMu.Unlock()
		exitCode = code
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

	go func() {
		flow.Main([]string{"task"})
	}()

	// Wait for the task to start and be blocked
	time.Sleep(100 * time.Millisecond)

	process, _ := os.FindProcess(os.Getpid())
	// Send two interrupt signals
	_ = process.Signal(os.Interrupt)
	time.Sleep(100 * time.Millisecond)
	_ = process.Signal(os.Interrupt)

	// Wait for the hard exit
	time.Sleep(100 * time.Millisecond)

	exitMu.Lock()
	gotCode := exitCode
	exitMu.Unlock()

	if gotCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", gotCode, exitCodeFail)
	}

	gotOutput := out.String()
	if !strings.Contains(gotOutput, "second interrupt, exit") {
		t.Errorf("output should contain hard exit message, got: %q", gotOutput)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	var exitMu sync.Mutex
	osExit = func(code int) {
		exitMu.Lock()
		defer exitMu.Unlock()
		exitCode = code
	}

	out := &safeBuffer{}
	SetOutput(out)
	defer SetOutput(nil)

	taskCanFinish := make(chan struct{})
	Define(Task{
		Name: "task",
		Action: func(_ *A) {
			<-taskCanFinish
		},
	})
	defer Undefine(Tasks()[0])

	done := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(done)
	}()

	// Wait for the task to start and be blocked
	time.Sleep(100 * time.Millisecond)

	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(os.Interrupt)

	time.Sleep(100 * time.Millisecond)
	close(taskCanFinish)

	<-done

	exitMu.Lock()
	gotCode := exitCode
	exitMu.Unlock()

	if gotCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", gotCode, exitCodeFail)
	}

	gotOutput := out.String()
	if !strings.Contains(gotOutput, "first interrupt, graceful stop") {
		t.Errorf("output should contain graceful stop message, got: %q", gotOutput)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCode int
	var exitMu sync.Mutex
	osExit = func(code int) {
		exitMu.Lock()
		defer exitMu.Unlock()
		exitCode = code
	}

	flow := &Flow{}
	flow.SetOutput(io.Discard)
	flow.Define(Task{Name: "task"})

	flow.Main([]string{"task"})

	exitMu.Lock()
	gotCode := exitCode
	exitMu.Unlock()

	if gotCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", gotCode, exitCodePass)
	}
}
