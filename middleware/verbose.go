package middleware

import (
	"io"
	"strings"

	"github.com/goyek/goyek/v2"
)

// SilentNonFailed banana.
func SilentNonFailed(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		orginalOut := in.Output
		streamWriter := &strings.Builder{}
		in.Output = streamWriter

		result := next(in)

		if result.Status == goyek.StatusFailed {
			io.Copy(orginalOut, strings.NewReader(streamWriter.String())) //nolint:errcheck,gosec // not checking errors when writing to output
		}

		return result
	}
}
