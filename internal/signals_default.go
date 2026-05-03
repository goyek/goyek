// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package internal

import (
	"os"
)

// TerminationSignals are the signals that should cause a graceful shutdown.
var TerminationSignals = []os.Signal{os.Interrupt}
