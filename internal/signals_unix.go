//go:build !windows && !plan9
// +build !windows,!plan9

package internal

import (
	"os"
	"syscall"
)

// TerminationSignals are signals that cause the program to terminate.
var TerminationSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
