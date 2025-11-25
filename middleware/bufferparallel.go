package middleware

import (
	"io"
	"strings"

	"github.com/goyek/goyek/v3"
)

// BufferParallel is a middleware which buffers the output from parallel tasks
// to not have mixed output from parallel tasks execution.
func BufferParallel(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		if !in.Parallel {
			return next(in)
		}

		orginalOut := in.Output
		streamWriter := &strings.Builder{}
		in.Output = streamWriter

		result := next(in)
		io.Copy(orginalOut, strings.NewReader(streamWriter.String())) //nolint:errcheck // not checking errors when writing to output
		return result
	}
}
