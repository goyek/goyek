package middleware

import (
	"io"
	"reflect"
)

var discardOutputType = reflect.TypeOf(io.Discard)

func outputOrDiscard(out io.Writer) io.Writer {
	if out == nil {
		return io.Discard
	}
	return out
}

func outputIsDiscard(out io.Writer) bool {
	return reflect.TypeOf(out) == discardOutputType
}
