//go:build windows || plan9
// +build windows plan9

package internal

import (
	"os"
)

// TerminationSignals are signals that cause the program to terminate.
var TerminationSignals = []os.Signal{os.Interrupt}
