//go:build windows || plan9
// +build windows plan9

package internal

import (
	"os"
)

// TerminationSignals returns the signals that should cause a graceful termination.
func TerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
