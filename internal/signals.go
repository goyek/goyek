package internal

import (
	"os"
)

// TerminationSignals returns the signals that cause the program to terminate.
func TerminationSignals() []os.Signal {
	signals := []os.Signal{os.Interrupt}
	for _, s := range platformSignals() {
		signals = append(signals, s)
	}
	return signals
}
