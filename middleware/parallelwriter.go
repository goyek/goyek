package middleware

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	parallelOutputFrame     = "=== NAME  "
	parallelTruncatedMarker = "\n\t... [output truncated]\n"
)

// parallelWriter line-buffers output so the task frame is never inserted in
// the middle of a logical line. Every framed record is sent to output with one
// Write call so any concurrency-safe output keeps the frame and line together.
type parallelWriter struct {
	mu         sync.Mutex
	output     io.Writer
	record     []byte
	prefixLen  int
	discarding bool
	closed     bool
	writeErr   error
}

func newParallelWriter(output io.Writer, taskName string) *parallelWriter {
	quotedTaskName := strconv.QuoteToGraphic(taskName)
	taskName = quotedTaskName[1 : len(quotedTaskName)-1]
	prefix := parallelOutputFrame + taskName + "\n"
	return &parallelWriter{
		output:    output,
		record:    append([]byte(nil), prefix...),
		prefixLen: len(prefix),
	}
}

func (w *parallelWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return w.writeBytes(p)
}

func (w *parallelWriter) WriteString(s string) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return w.writeString(s)
}

// ReadFrom preserves line buffering when this writer is the destination of
// io.Copy. Hiding r's WriterTo method keeps the copy buffer bounded even for a
// reader whose WriterTo implementation performs one enormous Write.
func (w *parallelWriter) ReadFrom(r io.Reader) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	if w.writeErr != nil {
		return 0, w.writeErr
	}

	destination := writerFunc(w.writeBytes)
	return io.Copy(destination, struct{ io.Reader }{r})
}

func (w *parallelWriter) writeBytes(p []byte) (int, error) {
	return w.writeInput(&parallelInput{bytes: p})
}

func (w *parallelWriter) writeString(s string) (int, error) {
	return w.writeInput(&parallelInput{text: s, isString: true})
}

func (w *parallelWriter) writeInput(input *parallelInput) (int, error) {
	written := 0
	currentRecordBytes := 0
	for input.length() > 0 {
		if w.discarding {
			newline := input.indexByte('\n')
			if newline < 0 {
				return written + input.length(), nil
			}
			consumed := newline + 1
			written += consumed
			input.consume(consumed)
			w.discarding = false
			continue
		}

		remaining := maxBufferedOutputBytes - w.payloadLen()
		newline := input.indexByte('\n')
		if newline >= 0 && newline <= remaining {
			lineLen := newline + 1
			w.record = input.appendTo(w.record, lineLen)
			currentRecordBytes += lineLen
			n, _, err := w.emit(currentRecordBytes, false)
			written += n
			if err != nil {
				return written, err
			}
			currentRecordBytes = 0
			input.consume(lineLen)
			continue
		}

		if remaining > input.length() {
			remaining = input.length()
		}
		w.record = input.appendTo(w.record, remaining)
		currentRecordBytes += remaining
		input.consume(remaining)
		if w.payloadLen() < maxBufferedOutputBytes || input.length() == 0 {
			return written + currentRecordBytes, nil
		}

		n, complete, err := w.emit(currentRecordBytes, true)
		if complete {
			w.discarding = true
		}
		written += n
		if err != nil {
			return written, err
		}
		currentRecordBytes = 0
	}
	return written + currentRecordBytes, nil
}

type parallelInput struct {
	bytes    []byte
	text     string
	isString bool
}

func (in *parallelInput) length() int {
	if in.isString {
		return len(in.text)
	}
	return len(in.bytes)
}

func (in *parallelInput) indexByte(value byte) int {
	if in.isString {
		return strings.IndexByte(in.text, value)
	}
	return bytes.IndexByte(in.bytes, value)
}

func (in *parallelInput) appendTo(destination []byte, n int) []byte {
	if in.isString {
		return append(destination, in.text[:n]...)
	}
	return append(destination, in.bytes[:n]...)
}

func (in *parallelInput) consume(n int) {
	if in.isString {
		in.text = in.text[n:]
		return
	}
	in.bytes = in.bytes[n:]
}

func (w *parallelWriter) emit(currentRecordBytes int, truncated bool) (int, bool, error) {
	payloadLen := w.payloadLen()
	previousRecordBytes := payloadLen - currentRecordBytes
	if truncated {
		payload := trimIncompleteUTF8(w.record[w.prefixLen:])
		w.record = w.record[:w.prefixLen+len(payload)]
		payloadLen = len(payload)
		if previousRecordBytes > payloadLen {
			previousRecordBytes = payloadLen
		}
		w.record = append(w.record, parallelTruncatedMarker...)
	}

	n, err := w.output.Write(w.record)
	complete := n == len(w.record)
	if !complete && err == nil {
		err = io.ErrShortWrite
	}
	if n < 0 {
		n = 0
	}
	if n > len(w.record) {
		n = len(w.record)
	}

	currentWritten := n - w.prefixLen - previousRecordBytes
	if currentWritten < 0 {
		currentWritten = 0
	}
	retainedCurrentBytes := payloadLen - previousRecordBytes
	if currentWritten > retainedCurrentBytes {
		currentWritten = retainedCurrentBytes
	}
	if complete {
		currentWritten = currentRecordBytes
	} else {
		w.writeErr = err
	}
	w.record = w.record[:w.prefixLen]
	return currentWritten, complete, err
}

func (w *parallelWriter) payloadLen() int {
	return len(w.record) - w.prefixLen
}

func trimIncompleteUTF8(p []byte) []byte {
	start := len(p) - 1
	min := len(p) - utf8.UTFMax
	if min < 0 {
		min = 0
	}
	for start >= min && !utf8.RuneStart(p[start]) {
		start--
	}
	if start >= min && !utf8.FullRune(p[start:]) {
		return p[:start]
	}
	return p
}

func (w *parallelWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return io.ErrClosedPipe
	}
	if w.writeErr != nil {
		return w.writeErr
	}
	return w.flush()
}

func (w *parallelWriter) flush() error {
	if w.discarding || w.payloadLen() == 0 {
		return nil
	}
	w.record = append(w.record, '\n')
	_, _, err := w.emit(0, false)
	return err
}

func (w *parallelWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}

	err := w.writeErr
	if err == nil {
		err = w.flush()
	}
	w.closed = true
	return err
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}

var (
	_ io.Writer       = (*parallelWriter)(nil)
	_ io.StringWriter = (*parallelWriter)(nil)
	_ io.ReaderFrom   = (*parallelWriter)(nil)
	_ io.Closer       = (*parallelWriter)(nil)
)
