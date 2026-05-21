package goyek

import (
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
)

const windows = "windows"

func TestFlow_Main_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}
	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	restoreTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = restoreTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})

	done := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(done)
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-done

	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}
	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}
	restoreTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = restoreTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() { close(hookCalled) }
	restoreTrapSignalsSecondHook := trapSignalsSecondHook
	defer func() { trapSignalsSecondHook = restoreTrapSignalsSecondHook }()
	secondHookCalled := make(chan struct{})
	trapSignalsSecondHook = func() { close(secondHookCalled) }
	f := &Flow{}
	f.SetOutput(io.Discard)
	taskCanFinish := make(chan struct{})
	f.Define(Task{Name: "task", Action: func(_ *A) { <-taskCanFinish }})
	done := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(done)
	}()
	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		runtime.Gosched()
	}
	<-secondHookCalled
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	close(taskCanFinish)
	<-done
	mu.Lock()
	code := exitCode
	mu.Unlock()
	if code != exitCodeFail {
		t.Errorf("got exit code %d, want %d", code, exitCodeFail)
	}
}

func TestFlow_Main_signal_hard_timeout(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}
	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	var mu sync.Mutex
	osExit = func(code int) {
		mu.Lock()
		exitCode = code
		mu.Unlock()
	}
	restoreTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = restoreTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() { close(hookCalled) }
	restoreTrapSignalsSecondHook := trapSignalsSecondHook
	defer func() { trapSignalsSecondHook = restoreTrapSignalsSecondHook }()
	secondHookCalled := make(chan struct{})
	trapSignalsSecondHook = func() { close(secondHookCalled) }
	f := &Flow{}
	f.SetOutput(io.Discard)
	taskCanFinish := make(chan struct{})
	f.Define(Task{Name: "task", Action: func(_ *A) { <-taskCanFinish }})
	done := make(chan struct{})
	go func() {
		f.Main([]string{"task"})
		close(done)
	}()
	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		runtime.Gosched()
	}
	<-secondHookCalled
	_ = p.Signal(os.Interrupt)
	_ = p.Signal(os.Interrupt)
	_ = p.Signal(os.Interrupt)
	close(taskCanFinish)
	<-done
	mu.Lock()
	code := exitCode
	mu.Unlock()
	if code != exitCodeFail {
		t.Errorf("got exit code %d, want %d", code, exitCodeFail)
	}
}

func TestFlow_Main_default_hooks(_ *testing.T) {
	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	osExit = func(int) {}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task"})
	f.Main([]string{"task"})
}

func TestMain_signal_graceful(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping signal test on windows")
	}
	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	restoreTrapSignalsHook := trapSignalsHook
	defer func() { trapSignalsHook = restoreTrapSignalsHook }()
	hookCalled := make(chan struct{})
	trapSignalsHook = func() {
		close(hookCalled)
	}

	f := DefaultFlow
	oldTasks := f.tasks
	f.tasks = make(map[string]*DefinedTask)
	defer func() { f.tasks = oldTasks }()

	Define(Task{
		Name: "task",
		Action: func(a *A) {
			<-a.Context().Done()
		},
	})

	done := make(chan struct{})
	go func() {
		Main([]string{"task"})
		close(done)
	}()

	<-hookCalled
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	<-done

	if exitCode != exitCodeFail {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodeFail)
	}
}

func TestFlow_Main_pass(t *testing.T) {
	restoreOsExit := osExit
	defer func() { osExit = restoreOsExit }()
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	f := &Flow{}
	f.SetOutput(io.Discard)
	f.Define(Task{Name: "task"})

	f.Main([]string{"task"})

	if exitCode != exitCodePass {
		t.Errorf("got exit code %d, want %d", exitCode, exitCodePass)
	}
}
