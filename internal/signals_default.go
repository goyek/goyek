//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package internal

import "os"

func init() {
	TerminationSignals = []os.Signal{os.Interrupt}
}
