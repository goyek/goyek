package internal

import "os"

// TerminationSignals returns the signals that should cause the program to terminate.
//
// On Unix-like systems, it returns [os.Interrupt] and [syscall.SIGTERM].
// On other systems, it returns [os.Interrupt].
func TerminationSignals() []os.Signal {
	return terminationSignals()
}
