package internal

import "os"

// TerminationSignals returns the signals that should trigger a graceful shutdown.
func TerminationSignals() []os.Signal {
	return terminationSignals()
}
