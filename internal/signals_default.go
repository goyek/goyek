//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris)
// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package internal

import "os"

func terminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
