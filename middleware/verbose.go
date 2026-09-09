package middleware

import (
	"io"
	"sync"

	"github.com/goyek/goyek/v3"
)

// SilentNonFailed is a middleware which makes sure that only output from failed tasks is printed.
//
// The behavior is based on the Go test runner when it is executed without the -v flag.
// It retains up to 1 MiB of output per task in memory, then spills the complete
// output to a permission-restricted file in the system temporary directory.
// Failed output is emitted in full, and output from other task results is
// discarded. The file is removed on a best-effort basis. Writes may return
// temporary file I/O errors after the in-memory limit is reached.
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
