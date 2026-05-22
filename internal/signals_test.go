package internal_test

import (
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/goyek/goyek/v3/internal"
)

func TestTerminationSignals(t *testing.T) {
	got := internal.TerminationSignals()

	switch runtime.GOOS {
	case "windows", "plan9":
		if len(got) != 1 || got[0] != os.Interrupt {
			t.Errorf("expected [os.Interrupt], got %v", got)
		}
	default:
		if len(got) != 2 || got[0] != os.Interrupt || got[1] != syscall.SIGTERM {
			t.Errorf("expected [os.Interrupt, syscall.SIGTERM], got %v", got)
		}
	}
}
