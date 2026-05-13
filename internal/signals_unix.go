//go:build !windows && !plan9
// +build !windows,!plan9

package internal

import (
	"os"
	"syscall"
)

func terminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
