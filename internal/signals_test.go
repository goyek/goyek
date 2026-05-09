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
		t.Fatal("TerminationSignals should not be empty")
	}

	foundInterrupt := false
	for _, sig := range sigs {
		if sig == os.Interrupt {
			foundInterrupt = true
			break
		}
	}
	if !foundInterrupt {
		t.Error("os.Interrupt should be in TerminationSignals")
	}

	switch runtime.GOOS {
	case "windows", "plan9":
		if len(sigs) != 1 {
			t.Errorf("got %d signals, want 1 on %s", len(sigs), runtime.GOOS)
		}
	default:
		if len(sigs) < 2 {
			t.Errorf("got %d signals, want at least 2 on %s", len(sigs), runtime.GOOS)
		}
	}
}
