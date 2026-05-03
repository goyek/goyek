package goyek

import (
	"context"
	"io"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestFlow_trapSignals(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows because os.Process.Signal(os.Interrupt) is not supported")
	}
	f := &Flow{}
	f.SetOutput(io.Discard)
	ctx, cancel := context.WithCancel(context.Background())

	f.trapSignals(cancel)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("could not find process: %v", err)
	}

	// Send signal to ourselves.
	// Since we are using a buffered channel of size 1 in trapSignals,
	// and we only care about the first signal for this test.
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("could not send signal: %v", err)
	}

	select {
	case <-ctx.Done():
		// Success
	case <-time.After(time.Second):
		t.Error("context should have been canceled")
	}
}
