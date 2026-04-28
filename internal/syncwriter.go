package internal

import (
	"io"
	"strings"
	"sync"
)

// SyncBuilder is a thread-safe strings.Builder.
type SyncBuilder struct {
	mu sync.Mutex
	sb strings.Builder
}

// Write appends the contents of p to b's buffer.
func (b *SyncBuilder) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.Write(p)
}

// WriteString appends the contents of s to b's buffer.
func (b *SyncBuilder) WriteString(s string) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.WriteString(s)
}

// String returns the accumulated string.
func (b *SyncBuilder) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.String()
}

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
