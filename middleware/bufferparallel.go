package middleware

import (
	"io"
	"sync"

	"github.com/goyek/goyek/v3"
)

// BufferParallel is a middleware which buffers the output from parallel tasks
// to not have mixed output from parallel tasks execution.
// It retains up to 1 MiB of output per task in memory, then spills the complete
// output to a permission-restricted file in the system temporary directory.
// Output is emitted in full after the task finishes, and the file is removed
// on a best-effort basis.
// Writes may return temporary file I/O errors after the in-memory limit is
// reached.
//
// Concurrent invocations through the same BufferParallel-wrapped runner
// serialize their complete replays. [goyek.Flow] additionally coordinates a
// replay with other output routed through the Flow. Outside a Flow, callers
// must share one [goyek.SyncWriter] result when the destination is also written
// through another runner or directly by other code.
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
