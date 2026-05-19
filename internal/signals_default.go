//go:build windows || plan9
// +build windows plan9

package internal

import (
	"os"
)

// TerminationSignals returns the signals that should cause the program to terminate.
func TerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
