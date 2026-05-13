package internal

import (
	"os"
	"runtime"
	"testing"
)

func TestTerminationSignals(t *testing.T) {
	sigs := TerminationSignals()

	if len(sigs) == 0 {
		t.Fatal("TerminationSignals() returned no signals")
	}

	hasInterrupt := false
	for _, sig := range sigs {
		if sig == os.Interrupt {
			hasInterrupt = true
			break
		}
	}
	if !hasInterrupt {
		t.Error("TerminationSignals() should include os.Interrupt")
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
