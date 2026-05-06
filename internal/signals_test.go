package internal_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/goyek/goyek/v3/internal"
)

func TestTerminationSignals(t *testing.T) {
	signals := internal.TerminationSignals()
	if len(signals) == 0 {
		t.Fatal("TerminationSignals should not be empty")
	}

	foundInterrupt := false
	for _, s := range signals {
		if s == os.Interrupt {
			foundInterrupt = true
			break
		}
	}
	if !foundInterrupt {
		t.Error("os.Interrupt not found in TerminationSignals")
	}

	switch runtime.GOOS {
	case "windows", "plan9":
		if len(signals) != 1 {
			t.Errorf("expected 1 signal on %s, got %d", runtime.GOOS, len(signals))
		}
	default:
		if len(signals) < 2 {
			t.Errorf("expected at least 2 signals on %s, got %d", runtime.GOOS, len(signals))
		}
	}
}
