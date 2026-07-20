package middleware_test

import (
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/middleware"
)

func TestReportLongRun(t *testing.T) {
	taskName := "my-task"
	sb := &strings.Builder{}
	r := goyek.NewRunner(func(*goyek.A) { time.Sleep(30 * time.Millisecond) })
	r = middleware.ReportLongRun(time.Millisecond)(r)

	r(goyek.Input{TaskName: taskName, Output: sb})

	if !strings.Contains(sb.String(), "***** LONG: "+taskName+" (") {
		t.Errorf("got: %q; but should long running report", sb.String())
	}
}

func TestReportLongRun_serializesTaskOutput(t *testing.T) {
	const taskOutput = "task output\n"
	out := newOverlapWriter(taskOutput)
	r := goyek.NewRunner(func(a *goyek.A) {
		io.WriteString(a.Output(), taskOutput) //nolint:errcheck // test writer does not return errors
	})
	r = middleware.ReportLongRun(time.Millisecond)(r)

	done := make(chan goyek.Result)
	go func() {
		done <- r(goyek.Input{TaskName: "task", Output: out})
	}()

	select {
	case <-out.taskWriteStarted:
	case <-time.After(time.Second):
		t.Fatal("task output did not start")
	}

	var overlapped bool
	select {
	case <-out.overlap:
		overlapped = true
	case <-time.After(100 * time.Millisecond):
	}
	close(out.releaseTaskWrite)

	select {
	case result := <-done:
		if result.Status != goyek.StatusPassed {
			t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
		}
	case <-time.After(time.Second):
		t.Fatal("runner did not finish")
	}

	if overlapped {
		t.Fatal("ReportLongRun and the task wrote to the underlying output concurrently")
	}
	if got := out.String(); !strings.Contains(got, taskOutput) || !strings.Contains(got, "***** LONG: task (") {
		t.Fatalf("expected task and long-running output, got %q", got)
	}
}

func TestReportLongRun_nilOutput(t *testing.T) {
	var gotOutput io.Writer
	r := middleware.ReportLongRun(time.Hour)(func(in goyek.Input) goyek.Result {
		gotOutput = in.Output
		return goyek.Result{Status: goyek.StatusPassed}
	})

	result := r(goyek.Input{})

	if result.Status != goyek.StatusPassed {
		t.Fatalf("got status %v, want %v", result.Status, goyek.StatusPassed)
	}
	if gotOutput == nil {
		t.Fatal("next runner received a nil output")
	}
}

func TestReportLongRun_stopsReporterWhenNextPanics(t *testing.T) {
	const panicValue = "runner panic"
	out := &notifyingWriter{written: make(chan struct{}, 1)}
	r := middleware.ReportLongRun(time.Millisecond)(func(goyek.Input) goyek.Result {
		select {
		case <-out.written:
		case <-time.After(time.Second):
			t.Fatal("long-running report was not written")
		}
		panic(panicValue)
	})

	var recovered interface{}
	func() {
		defer func() {
			recovered = recover()
		}()
		r(goyek.Input{TaskName: "task", Output: out})
	}()

	if recovered != panicValue {
		t.Fatalf("recovered %v, want %q", recovered, panicValue)
	}
	writes := atomic.LoadInt32(&out.writes)
	time.Sleep(10 * time.Millisecond)
	if got := atomic.LoadInt32(&out.writes); got != writes {
		t.Fatalf("long-running reporter continued after panic: writes changed from %d to %d", writes, got)
	}
}

func TestReportLongRun_wrapsOutputBeforeNext(t *testing.T) {
	out := &strings.Builder{}
	var gotOutput io.Writer
	r := middleware.ReportLongRun(time.Hour)(func(in goyek.Input) goyek.Result {
		gotOutput = in.Output
		return goyek.Result{Status: goyek.StatusPassed}
	})

	r(goyek.Input{Output: out})

	if gotOutput == out {
		t.Fatal("next runner received the unsynchronized output")
	}
}

type overlapWriter struct {
	active           int32
	overlap          chan struct{}
	taskOutput       string
	taskWriteStarted chan struct{}
	releaseTaskWrite chan struct{}
	taskWriteOnce    sync.Once
	mu               sync.Mutex
	buf              strings.Builder
}

func newOverlapWriter(taskOutput string) *overlapWriter {
	return &overlapWriter{
		overlap:          make(chan struct{}, 1),
		taskOutput:       taskOutput,
		taskWriteStarted: make(chan struct{}),
		releaseTaskWrite: make(chan struct{}),
	}
}

func (w *overlapWriter) Write(p []byte) (int, error) {
	if atomic.AddInt32(&w.active, 1) > 1 {
		select {
		case w.overlap <- struct{}{}:
		default:
		}
	}
	defer atomic.AddInt32(&w.active, -1)

	if string(p) == w.taskOutput {
		w.taskWriteOnce.Do(func() {
			close(w.taskWriteStarted)
			<-w.releaseTaskWrite
		})
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *overlapWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

type notifyingWriter struct {
	writes  int32
	written chan struct{}
}

func (w *notifyingWriter) Write(p []byte) (int, error) {
	atomic.AddInt32(&w.writes, 1)
	select {
	case w.written <- struct{}{}:
	default:
	}
	return len(p), nil
}
