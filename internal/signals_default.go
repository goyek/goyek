//go:build windows || plan9
// +build windows plan9

package internal

import (
	"os"
)

// TerminationSignals returns the signals that should cause a graceful shutdown.
func TerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
