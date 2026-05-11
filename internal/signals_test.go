package internal

import (
	"os"
	"testing"
)

func TestTerminationSignals(t *testing.T) {
	sigs := TerminationSignals()
	if len(sigs) == 0 {
		t.Fatal("TerminationSignals() returned no signals")
	}
	foundInterrupt := false
	for _, sig := range sigs {
		if sig == os.Interrupt {
			foundInterrupt = true
			break
		}
	}
	if !foundInterrupt {
		t.Error("TerminationSignals() did not return os.Interrupt")
	}
}
