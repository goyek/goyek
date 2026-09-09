package middleware

import (
	"io"
	"sync"

	"github.com/goyek/goyek/v3"
)

// SilentNonFailed is a middleware which buffers task output and emits it only
// when the task fails. Output from other task results is discarded.
//
// The behavior is based on the Go test runner when it is executed without the -v flag.
func SilentNonFailed(next goyek.Runner) goyek.Runner {
	var replayMu sync.Mutex

	return func(in goyek.Input) goyek.Result {
		originalOut := outputOrDiscard(in.Output)
		if outputIsDiscard(originalOut) {
			in.Output = io.Discard
			return next(in)
		}

		streamWriter := newSpoolBuffer(maxBufferedOutputBytes)
		defer func() {
			_ = streamWriter.Close()
		}()
		in.Output = goyek.SyncWriter(streamWriter)

		result := next(in)

		if result.Status == goyek.StatusFailed {
			func() {
				replayMu.Lock()
				defer replayMu.Unlock()
				streamWriter.WriteTo(originalOut) //nolint:errcheck // not checking errors when writing to output
			}()
		}

		return result
	}
}
