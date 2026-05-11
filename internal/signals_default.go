//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package internal

import (
	"os"
)

// TerminationSignals returns the signals that should cause the program to terminate.
func TerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
