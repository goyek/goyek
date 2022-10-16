package goyek

import (
	"io"
	"strings"
)

func silentNonFailed(next runner) runner {
	return func(in input) result {
		orginalOut := in.Output
		streamWriter := &strings.Builder{}
		in.Output = streamWriter

		result := next(in)

		if result.status == statusFailed || result.status == statusPanicked {
			io.Copy(orginalOut, strings.NewReader(streamWriter.String())) //nolint:errcheck,gosec // not checking errors when writing to output
		}

		return result
	}
}
