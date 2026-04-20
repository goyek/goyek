package goyek

import (
	"io"
	"sync"
)

// SyncWriter is a thread-safe [io.Writer] and [io.StringWriter] wrapper.
type SyncWriter struct {
	io.Writer
	mtx sync.Mutex
}

// Write implements [io.Writer].
func (w *SyncWriter) Write(p []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.Writer.Write(p)
}

// WriteString implements [io.StringWriter].
func (w *SyncWriter) WriteString(s string) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return io.WriteString(w.Writer, s)
}
