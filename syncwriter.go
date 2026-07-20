package goyek

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

// SyncWriter returns a writer safe for concurrent use by serializing calls to
// Write and WriteString. Each call is a synchronization unit; a logical record
// split across multiple calls may be interleaved with other writes. The returned
// writer implements [io.StringWriter], but does not preserve other optional
// interfaces implemented by w.
// It returns nil if w is nil.
//
// Calling SyncWriter with a writer previously returned by SyncWriter returns
// that writer unchanged. Call SyncWriter once and share its result: separately
// wrapping the same underlying writer creates independent locks and does not
// make access through those wrappers safe.
//
// Access to w through any other reference, including reads of its state, is not
// synchronized with calls made through the returned writer. Do not access w
// directly until those calls have finished unless w is independently safe for
// concurrent use.
//
// [Flow.Execute] and [Flow.Main] already serialize writes they route to their
// configured output. SyncWriter is useful for standalone runner or executor
// inputs, middleware replacements, and a destination shared across flows or
// with code outside a flow execution.
func SyncWriter(w io.Writer) io.Writer {
	if w == nil {
		return nil
	}
	if sw, ok := w.(*syncWriter); ok {
		return sw
	}
	return &syncWriter{writer: w}
}
