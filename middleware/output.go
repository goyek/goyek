package middleware

import "io"

func outputOrDiscard(out io.Writer) io.Writer {
	if out == nil {
		return io.Discard
	}
	return out
}
