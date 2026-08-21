package middleware

import (
	"io"

	"github.com/goyek/goyek/v3"
)

// BufferParallel is a middleware which streams line-buffered output from
// parallel tasks and frames each line with its task name so interleaved output
// can be attributed without retaining the entire task output in memory, like
// the verbose Go test runner.
// Lines longer than 1 MiB are truncated with a visible marker. A final
// incomplete line is terminated with a newline when the task finishes.
func BufferParallel(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		if !in.Parallel {
			return next(in)
		}
		if in.Output == nil || outputIsDiscard(in.Output) {
			in.Output = io.Discard
			return next(in)
		}

		originalOut := outputOrDiscard(in.Output)
		streamWriter := newParallelWriter(originalOut, in.TaskName)
		defer func() {
			_ = streamWriter.Close() // best-effort final partial-line flush
		}()
		in.Output = streamWriter
		return next(in)
	}
}
