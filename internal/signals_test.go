package internal_test

import (
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/goyek/goyek/v3/internal"
)

func TestTerminationSignals(t *testing.T) {
	hasInterrupt := false
	hasSIGTERM := false
	for _, s := range internal.TerminationSignals() {
		if s == os.Interrupt {
			hasInterrupt = true
		}
		if s == syscall.SIGTERM {
			hasSIGTERM = true
		}
	}

	if !hasInterrupt {
		t.Error("TerminationSignals should contain os.Interrupt")
	}

	isUnix := false
	switch runtime.GOOS {
	case "aix", "darwin", "dragonfly", "freebsd", "linux", "netbsd", "openbsd", "solaris":
		isUnix = true
	}

	if isUnix && !hasSIGTERM {
		t.Error("TerminationSignals should contain syscall.SIGTERM on Unix")
	}
	if !isUnix && hasSIGTERM {
		t.Error("TerminationSignals should not contain syscall.SIGTERM on non-Unix")
	}
}
