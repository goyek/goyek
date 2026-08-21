package middleware

import (
	"io"

	"github.com/goyek/goyek/v3"
)

// SilentNonFailed is a middleware which makes sure that only output from failed tasks is printed.
//
// Like the Go test runner without -v, it defers output until the result is known.
// It retains up to 1 MiB of output per task in memory, then spills additional
// output to a temporary file. It makes a best-effort attempt to remove the file
// when the task finishes. Writes may return temporary-file I/O errors after the
// in-memory limit is reached.
func SilentNonFailed(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		if in.Output == nil || outputIsDiscard(in.Output) {
			in.Output = io.Discard
			return next(in)
		}

		originalOut := outputOrDiscard(in.Output)
		streamWriter := newSpoolBuffer(maxBufferedOutputBytes)
		defer func() {
			_ = streamWriter.Close() // best-effort temporary file cleanup
		}()
		in.Output = goyek.SyncWriter(streamWriter)

		result := next(in)

		if result.Status == goyek.StatusFailed {
			streamWriter.WriteTo(originalOut) //nolint:errcheck // not checking errors when writing to output
		}

		return result
	}
}
