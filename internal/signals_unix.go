// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package internal

import (
	"os"
	"syscall"
)

// TerminationSignals are the signals that should cause a graceful shutdown.
var TerminationSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
