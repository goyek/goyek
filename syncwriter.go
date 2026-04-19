package goyek

import (
	"io"
	"sync"
)

// SyncWriter is a thread-safe [io.Writer] and [io.StringWriter].
type SyncWriter struct {
	io.Writer
	mu sync.Mutex
}

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
