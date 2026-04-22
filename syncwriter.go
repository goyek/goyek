package goyek

import (
	"io"
	"sync"
)

// SyncWriter is a thread-safe writer.
// It also implements [io.StringWriter].
type SyncWriter struct {
	// Writer is the underlying writer.
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
	return w.Writer.Write([]byte(s))
}

// Sync returns a thread-safe [io.Writer] and [io.StringWriter].
// If w is already a [*SyncWriter] it is returned as is.
// If w is nil, [io.Discard] is used.
func Sync(w io.Writer) *SyncWriter {
	if w == nil {
		w = io.Discard
	}
	if syncW, ok := w.(*SyncWriter); ok {
		return syncW
	}
	return &SyncWriter{Writer: w}
}
