//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package internal

import (
	"syscall"
)

func init() {
	TerminationSignals = append(TerminationSignals, syscall.SIGTERM)
}
