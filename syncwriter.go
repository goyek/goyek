package goyek

import (
	"io"
	"sync"
)

type syncWriter struct {
	io.Writer
	mtx sync.Mutex
}

func (w *syncWriter) Write(p []byte) (int, error) {
	defer func() { w.mtx.Unlock() }()
	w.mtx.Lock()
	return w.Writer.Write(p)
}

func synchronizeWriter(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	if syncW, ok := w.(*syncWriter); ok {
		return syncW
	}
	return &syncWriter{Writer: w}
}
