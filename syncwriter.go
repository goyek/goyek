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

// WriteString implements io.StringWriter.
func (w *syncWriter) WriteString(s string) (int, error) {
	defer func() { w.mtx.Unlock() }()
	w.mtx.Lock()
	if sw, ok := w.Writer.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	return w.Writer.Write([]byte(s))
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

// SyncWriter is a thread-safe [io.Writer] and [io.StringWriter].
type SyncWriter struct {
	io.Writer
}

// Write implements [io.Writer].
func (w SyncWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}

// WriteString implements [io.StringWriter].
func (w SyncWriter) WriteString(s string) (int, error) {
	if sw, ok := w.Writer.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	return w.Writer.Write([]byte(s))
}
