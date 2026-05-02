package goyek

import (
	"context"
	"io"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/goyek/goyek/v3/internal"
)

func TestFlow_main(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(&strings.Builder{})
	flow.Define(Task{Name: "task"})
	flow.Define(Task{Name: "failing", Action: func(a *A) { a.Fail() }})

	testCases := []struct {
		desc string
		want int
		act  func() int
	}{
		{
			desc: "pass",
			want: 0,
			act:  func() int { return flow.main(context.Background(), []string{"task"}) },
		},
		{
			desc: "fail",
			want: 1,
			act:  func() int { return flow.main(context.Background(), []string{"failing"}) },
		},
		{
			desc: "invalid",
			want: 2,
			act:  func() int { return flow.main(context.Background(), []string{"bad"}) },
		},
		{
			desc: "canceled",
			want: 1,
			act: func() int {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return flow.main(ctx, []string{"task"})
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.act(); got != tc.want {
				t.Errorf("got: %d; want: %d", got, tc.want)
			}
		})
	}
}

func Test_main_usage(t *testing.T) {
	flow := &Flow{}
	flow.SetOutput(io.Discard)
	called := false
	flow.SetUsage(func() { called = true })

	flow.main(context.Background(), nil)

	if !called {
		t.Error("usage should be called for invalid input")
	}
}

func TestFlow_Main(t *testing.T) {
	flow := &Flow{}
	sb := &safeBuffer{}
	flow.SetOutput(sb)
	flow.Define(Task{Name: "task"})
	flow.Define(Task{Name: "wait", Action: func(a *A) {
		<-a.Context().Done()
	}})

	exitCh := make(chan int, 1)
	mainExiterMu.Lock()
	origExiter := mainExiter
	mainExiter = func(code int) {
		select {
		case exitCh <- code:
		default:
		}
	}
	mainExiterMu.Unlock()
	defer func() {
		mainExiterMu.Lock()
		mainExiter = origExiter
		mainExiterMu.Unlock()
	}()

	t.Run("pass", func(t *testing.T) {
		flow.Main([]string{"task"})
		gotExitCode := <-exitCh
		if gotExitCode != 0 {
			t.Errorf("got exit code %d, want 0", gotExitCode)
		}
	})

	t.Run("interrupt", func(t *testing.T) {
		if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
			t.Skip("skipping on " + runtime.GOOS)
		}

		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), internal.TerminationSignals[0].(syscall.Signal))
		}()

		flow.Main([]string{"wait"})
		gotExitCode := <-exitCh
		if gotExitCode != 1 {
			t.Errorf("got exit code %d, want 1", gotExitCode)
		}
		if !strings.Contains(sb.String(), "first interrupt") {
			t.Error("missing first interrupt message")
		}
	})
}

func TestMain_wrapper(t *testing.T) {
	mainExiterMu.Lock()
	origExiter := mainExiter
	mainExiter = func(code int) {}
	mainExiterMu.Unlock()
	defer func() {
		mainExiterMu.Lock()
		mainExiter = origExiter
		mainExiterMu.Unlock()
	}()

	Main(nil)
}

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

func TestFailError_Error(t *testing.T) {
	err := &FailError{Task: "test"}
	want := "task failed: test"
	if got := err.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFlow_Main_second_interrupt(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("skipping on " + runtime.GOOS)
	}

	flow := &Flow{}
	sb := &safeBuffer{}
	flow.SetOutput(sb)
	flow.Define(Task{Name: "wait", Action: func(a *A) {
		<-a.Context().Done()
		time.Sleep(1 * time.Second)
	}})

	exitCh := make(chan int, 1)
	mainExiterMu.Lock()
	origExiter := mainExiter
	mainExiter = func(code int) {
		select {
		case exitCh <- code:
		default:
		}
	}
	mainExiterMu.Unlock()
	defer func() {
		mainExiterMu.Lock()
		mainExiter = origExiter
		mainExiterMu.Unlock()
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)

		for i := 0; i < 20; i++ {
			if strings.Contains(sb.String(), "first interrupt") {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	flow.Main([]string{"wait"})

	gotExitCode := <-exitCh
	if gotExitCode != 1 {
		t.Errorf("got exit code %d, want 1", gotExitCode)
	}
}
