package goyek

import (
	"io"
	"sync"
)

// SyncWriter is a thread-safe [io.Writer] and [io.StringWriter].
type SyncWriter struct {
	Writer io.Writer
	mu     sync.Mutex
}

// Write writes p to the underlying [io.Writer] in a thread-safe manner.
func (w *SyncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Writer.Write(p)
}

// WriteString writes s to the underlying [io.Writer] in a thread-safe manner.
// If the underlying writer implements [io.StringWriter], it is used.
func (w *SyncWriter) WriteString(s string) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if sw, ok := w.Writer.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	return w.Writer.Write([]byte(s))
}
