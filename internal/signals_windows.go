//go:build windows || plan9
// +build windows plan9

package internal

import "os"

func init() {
	TerminationSignals = []os.Signal{os.Interrupt}
}
