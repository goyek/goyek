//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package internal

import (
	"os"
	"syscall"
)

func init() {
	TerminationSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
}
