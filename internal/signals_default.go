//go:build windows || plan9
// +build windows plan9

package internal

import "os"

func terminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
