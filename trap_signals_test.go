package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

const windows = "windows"

func TestFlow_Main_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCodes []int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCodes = append(exitCodes, code)
	}

	origTrapSignalsHook := trapSignalsHook
	ready := make(chan struct{})
	trapSignalsHook = func() {
		close(ready)
	}
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(doneCh)
	}()

	<-ready
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if len(exitCodes) == 0 || exitCodes[len(exitCodes)-1] != exitCodeFail {
		t.Errorf("got exit codes %v, want last to be %d", exitCodes, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	osExitCalled := make(chan int, 10)
	osExit = func(code int) {
		osExitCalled <- code
	}

	origTrapSignalsHook := trapSignalsHook
	ready := make(chan struct{})
	trapSignalsHook = func() {
		close(ready)
	}
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	origTrapSignalsSecondHook := trapSignalsSecondHook
	secondReady := make(chan struct{})
	trapSignalsSecondHook = func() {
		close(secondReady)
	}
	defer func() { trapSignalsSecondHook = origTrapSignalsSecondHook }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
			// first signal received, now wait for second signal via osExitCalled
			// or timeout to avoid blocking forever if test fails
			select {
			case <-osExitCalled:
			case <-time.After(time.Second):
			}
		},
	})

	doneCh := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(doneCh)
	}()

	<-ready
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-secondReady
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh

	// The last exit code should be from Main's final call to osExit
	// The one before it should be exitCodeFail from the signal handler
	var codes []int
L:
	for {
		select {
		case c := <-osExitCalled:
			codes = append(codes, c)
		default:
			break L
		}
	}

	foundFail := false
	for _, c := range codes {
		if c == exitCodeFail {
			foundFail = true
			break
		}
	}
	if !foundFail {
		t.Errorf("got exit codes %v, want to find %d", codes, exitCodeFail)
	}
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}

	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCodes []int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCodes = append(exitCodes, code)
	}

	origTrapSignalsHook := trapSignalsHook
	ready := make(chan struct{})
	trapSignalsHook = func() {
		close(ready)
	}
	defer func() { trapSignalsHook = origTrapSignalsHook }()

	f := &Flow{}
	f.SetOutput(io.Discard)
	task := f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})
	f.SetDefault(task)

	oldDefaultFlow := DefaultFlow
	DefaultFlow = f
	defer func() { DefaultFlow = oldDefaultFlow }()

	doneCh := make(chan struct{})
	go func() {
		Main(nil)
		close(doneCh)
	}()

	<-ready
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-doneCh

	mu.Lock()
	defer mu.Unlock()
	if len(exitCodes) == 0 || exitCodes[len(exitCodes)-1] != exitCodeFail {
		t.Errorf("got exit codes %v, want last to be %d", exitCodes, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCodes []int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCodes = append(exitCodes, code)
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task"})

	f.Main([]string{"task"})

	mu.Lock()
	defer mu.Unlock()
	if len(exitCodes) == 0 || exitCodes[len(exitCodes)-1] != exitCodePass {
		t.Errorf("got exit codes %v, want last to be %d", exitCodes, exitCodePass)
	}
}

func TestFlow_Main_fail(t *testing.T) {
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	var exitCodes []int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCodes = append(exitCodes, code)
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task", Action: func(a *A) { a.Fail() }})

	f.Main([]string{"task"})

	mu.Lock()
	defer mu.Unlock()
	if len(exitCodes) == 0 || exitCodes[len(exitCodes)-1] != exitCodeFail {
		t.Errorf("got exit codes %v, want last to be %d", exitCodes, exitCodeFail)
	}
}

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "my-task"}
	want := "task failed: my-task"
	if got := err.Error(); got != want {
		t.Errorf("got: %q; want: %q", got, want)
	}
}
