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
	if sigs[0] != os.Interrupt {
		t.Errorf("expected first signal to be os.Interrupt, got %v", sigs[0])
	}
	if runtime.GOOS != "windows" && runtime.GOOS != "plan9" {
		if len(sigs) < 2 {
			t.Fatal("expected at least 2 termination signals on Unix")
		}
	}
}
