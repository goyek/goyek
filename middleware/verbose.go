package middleware

import (
	"io"
	"strings"

	"github.com/goyek/goyek/v3"
)

// SilentNonFailed is a middleware which makes sure that only output from failed tasks is printed.
//
// The behavior is based on the Go test runner when it is executed without the -v flag.
func SilentNonFailed(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		orginalOut := in.Output
		streamWriter := &strings.Builder{}
		in.Output = streamWriter

		result := next(in)

		if result.Status == goyek.StatusFailed {
			io.Copy(orginalOut, strings.NewReader(streamWriter.String())) //nolint:errcheck // not checking errors when writing to output
		}

		return result
	}
}
