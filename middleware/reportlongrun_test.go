package middleware_test

import (
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/middleware"
)

func TestReportLongRun(t *testing.T) {
	taskName := "my-task"
	sb := &strings.Builder{}
	out := goyek.SyncWriter(sb)
	r := goyek.NewRunner(func(*goyek.A) { time.Sleep(30 * time.Millisecond) })
	r = middleware.ReportLongRun(time.Millisecond)(r)

	r(goyek.Input{TaskName: taskName, Output: out})

	if !strings.Contains(sb.String(), "***** LONG: "+taskName+" (") {
		t.Errorf("got: %q; but should long running report", sb.String())
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
	if gotOutput != io.Discard {
		t.Fatalf("next runner received %T output, want io.Discard", gotOutput)
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

func TestReportLongRun_preservesOutput(t *testing.T) {
	out := io.Discard
	var gotOutput io.Writer
	r := middleware.ReportLongRun(time.Hour)(func(in goyek.Input) goyek.Result {
		gotOutput = in.Output
		return goyek.Result{Status: goyek.StatusPassed}
	})

	r(goyek.Input{Output: out})

	if gotOutput != out {
		t.Fatal("ReportLongRun replaced Input.Output")
	}
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
