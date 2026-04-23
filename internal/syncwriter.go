package internal

import (
	"io"
	"sync"
)

type syncWriter struct {
	writer io.Writer
	mu     sync.Mutex
}

func (w *syncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

func (w *syncWriter) WriteString(s string) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return io.WriteString(w.writer, s)
}

var _ io.StringWriter = (*syncWriter)(nil)

// SyncWriter synchronizes writes to the underlying writer.
func SyncWriter(w io.Writer) io.Writer {
	if sw, ok := w.(*syncWriter); ok {
		return sw
	}
	return &syncWriter{writer: w}
}
