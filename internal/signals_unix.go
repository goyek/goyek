//go:build !windows && !plan9
// +build !windows,!plan9

package internal

import (
	"os"
	"syscall"
)

// TerminationSignals returns the signals that cause the program to terminate.
func TerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
