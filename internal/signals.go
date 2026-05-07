package internal

import "os"

// TerminationSignals returns the signals that should cause a graceful termination.
func TerminationSignals() []os.Signal {
	return terminationSignals()
}
