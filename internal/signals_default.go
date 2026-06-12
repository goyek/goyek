//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package internal

import "os"

// TerminationSignals returns signals that should cause a graceful shutdown.
func TerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
