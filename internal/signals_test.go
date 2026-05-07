package internal_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/goyek/goyek/v3/internal"
)

func TestTerminationSignals(t *testing.T) {
	sigs := internal.TerminationSignals()

	if len(sigs) == 0 {
		t.Fatal("no termination signals returned")
	}

	hasInterrupt := false
	for _, sig := range sigs {
		if sig == os.Interrupt {
			hasInterrupt = true
			break
		}
	}
	if !hasInterrupt {
		t.Error("os.Interrupt missing from termination signals")
	}

	switch runtime.GOOS {
	case "windows", "plan9":
		if len(sigs) != 1 {
			t.Errorf("expected 1 signal on %s, got %d", runtime.GOOS, len(sigs))
		}
	default:
		if len(sigs) < 2 {
			t.Errorf("expected at least 2 signals on %s, got %d", runtime.GOOS, len(sigs))
		}
	}
}
