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
		t.Fatal("no termination signals")
	}

	foundInterrupt := false
	for _, sig := range sigs {
		if sig == os.Interrupt {
			foundInterrupt = true
			break
		}
	}
	if !foundInterrupt {
		t.Error("os.Interrupt not found in termination signals")
	}

	if runtime.GOOS != "windows" && runtime.GOOS != "plan9" {
		if len(sigs) < 2 {
			t.Errorf("expected at least 2 termination signals on %s, got %d", runtime.GOOS, len(sigs))
		}
	}
}
