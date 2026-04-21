package goyek

import (
	"io"
	"sync"
)

// SyncWriter is a thread-safe [io.Writer] and [io.StringWriter] wrapper.
type SyncWriter struct {
	Writer io.Writer
	mu     sync.Mutex
}

// Write implements [io.Writer].
func (w *SyncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Writer.Write(p)
}

// WriteString implements [io.StringWriter].
func (w *SyncWriter) WriteString(s string) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if sw, ok := w.Writer.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	return io.WriteString(w.Writer, s)
}

// Sync returns a [SyncWriter] that wraps w.
// It returns w if it is already a [*SyncWriter].
// If w is nil, it returns a [*SyncWriter] wrapping [io.Discard].
func Sync(w io.Writer) *SyncWriter {
	if w == nil {
		return &SyncWriter{Writer: io.Discard}
	}
	if syncW, ok := w.(*SyncWriter); ok {
		return syncW
	}
	return &SyncWriter{Writer: w}
}
