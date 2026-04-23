package goyek

import (
	"io"
	"sync"
)

// SyncWriter synchronizes writes to the underlying writer.
//
// It is used by [A] and middlewares to ensure that logs from multiple goroutines
// are not interleaved.
type SyncWriter struct {
	// Writer is the underlying writer.
	Writer io.Writer
	mu     sync.Mutex
}

// Write writes p to the underlying writer.
func (w *SyncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Writer.Write(p)
}

// WriteString writes s to the underlying writer.
func (w *SyncWriter) WriteString(s string) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if sw, ok := w.Writer.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	return io.WriteString(w.Writer, s)
}

// Sync synchronizes writes to the underlying writer.
// If the writer is already a *SyncWriter, it is returned as is.
func Sync(w io.Writer) *SyncWriter {
	if sw, ok := w.(*SyncWriter); ok {
		return sw
	}
	return &SyncWriter{Writer: w}
}
