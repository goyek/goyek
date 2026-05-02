package internal

import "os"

// TerminationSignals are the signals that should cause the flow to stop gracefully.
var TerminationSignals []os.Signal
