package middleware

import (
	"io"
	"sync"

	"github.com/goyek/goyek/v3"
)

// BufferParallel is a middleware which buffers output from parallel tasks to
// prevent it from mixing during parallel task execution. Each parallel task's
// complete output is emitted after the task finishes. Non-parallel tasks pass
// through without buffering.
func BufferParallel(next goyek.Runner) goyek.Runner {
	var replayMu sync.Mutex

	return func(in goyek.Input) goyek.Result {
		if !in.Parallel {
			return next(in)
		}

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
		func() {
			replayMu.Lock()
			defer replayMu.Unlock()
			streamWriter.WriteTo(originalOut) //nolint:errcheck // not checking errors when writing to output
		}()
		return result
	}
}
