//go:build !windows && !plan9
// +build !windows,!plan9

package internal

import (
	"os"
	"syscall"
)

func platformSignals() []os.Signal {
	return []os.Signal{syscall.SIGTERM}
}
