//go:build !windows && !plan9
// +build !windows,!plan9

package internal

import (
	"os"
	"syscall"
)

func init() {
	TerminationSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
}
