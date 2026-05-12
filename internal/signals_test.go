package internal_test

import (
	"os"
	"testing"

	"github.com/goyek/goyek/v3/internal"
)

func TestTerminationSignals(t *testing.T) {
	sigs := internal.TerminationSignals()
	if len(sigs) == 0 {
		t.Fatal("no signals returned")
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
}
