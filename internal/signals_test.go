package internal

import (
	"os"
	"runtime"
	"syscall"
	"testing"
)

func TestTerminationSignals(t *testing.T) {
	signals := make(map[os.Signal]bool)
	for _, s := range TerminationSignals {
		signals[s] = true
	}

	if !signals[os.Interrupt] {
		t.Error("os.Interrupt missing from TerminationSignals")
	}

	switch runtime.GOOS {
	case "windows", "plan9":
		if len(TerminationSignals) != 1 {
			t.Errorf("expected 1 signal on %s, got %d", runtime.GOOS, len(TerminationSignals))
		}
	default:
		if !signals[syscall.SIGTERM] {
			t.Error("syscall.SIGTERM missing from TerminationSignals")
		}
		if len(TerminationSignals) != 2 {
			t.Errorf("expected 2 signals on %s, got %d", runtime.GOOS, len(TerminationSignals))
		}
	}
}
