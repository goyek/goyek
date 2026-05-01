//go:build windows || plan9
// +build windows plan9

package internal

import (
	"os"
)

func platformSignals() []os.Signal {
	return nil
}
