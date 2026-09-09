package middleware

import (
	"bytes"
	"errors"
	"io"
	"os"
)

const maxBufferedOutputBytes = 1 << 20

var errSpoolBufferClosed = errors.New("middleware: spool buffer is closed")

type spoolFile interface {
	io.Reader
	io.Writer
	io.StringWriter
	io.Seeker
	io.Closer
	Name() string
}

type spoolBuffer struct {
	limit        int
	memory       *bytes.Buffer
	file         spoolFile
	createTemp   func() (spoolFile, error)
	remove       func(string) error
	size         int64
	replayOffset int64
	closed       bool
}

func newSpoolBuffer(limit int) *spoolBuffer {
	return newSpoolBufferWithFileOps(
		limit,
		func() (spoolFile, error) {
			return os.CreateTemp("", "goyek-output-")
		},
		os.Remove,
	)
}

func newSpoolBufferWithFileOps(
	limit int,
	createTemp func() (spoolFile, error),
	remove func(string) error,
) *spoolBuffer {
	if limit < 0 {
		limit = 0
	}
	return &spoolBuffer{
		limit:      limit,
		memory:     &bytes.Buffer{},
		createTemp: createTemp,
		remove:     remove,
	}
}

func (b *spoolBuffer) Write(p []byte) (int, error) {
	if b.closed {
		return 0, errSpoolBufferClosed
	}
	if len(p) == 0 {
		return 0, nil
	}
	if b.file != nil {
		return b.writeFile(p)
	}
	if len(p) <= b.limit-b.memory.Len() {
		return b.memory.Write(p)
	}
	if err := b.spill(); err != nil {
		return 0, err
	}
	return b.writeFile(p)
}

func (b *spoolBuffer) WriteString(s string) (int, error) {
	if b.closed {
		return 0, errSpoolBufferClosed
	}
	if len(s) == 0 {
		return 0, nil
	}
	if b.file != nil {
		return b.writeFileString(s)
	}
	if len(s) <= b.limit-b.memory.Len() {
		return b.memory.WriteString(s)
	}
	if err := b.spill(); err != nil {
		return 0, err
	}
	return b.writeFileString(s)
}

func (b *spoolBuffer) WriteTo(w io.Writer) (int64, error) {
	if b.closed {
		return 0, errSpoolBufferClosed
	}
	if b.file == nil {
		return b.memory.WriteTo(w)
	}
	if b.replayOffset == b.size {
		return 0, nil
	}
	if _, err := b.file.Seek(b.replayOffset, io.SeekStart); err != nil {
		return 0, err
	}

	remaining := b.size - b.replayOffset
	reader := &io.LimitedReader{
		R: readerOnly{Reader: b.file},
		N: remaining,
	}
	n, err := io.Copy(w, reader)
	b.replayOffset += n
	if err == nil && n != remaining {
		err = io.ErrUnexpectedEOF
	}
	return n, err
}

func (b *spoolBuffer) Close() error {
	if b.closed {
		return nil
	}
	b.closed = true
	b.memory = nil
	if b.file == nil {
		return nil
	}

	name := b.file.Name()
	closeErr := b.file.Close()
	b.file = nil
	removeErr := b.remove(name)
	if closeErr != nil {
		return closeErr
	}
	return removeErr
}

func (b *spoolBuffer) spill() error {
	file, err := b.createTemp()
	if err != nil {
		return err
	}

	n, err := writeOnce(file, b.memory.Bytes())
	if err != nil {
		_ = file.Close()
		_ = b.remove(file.Name())
		return err
	}

	b.file = file
	b.size = int64(n)
	b.memory = nil
	return nil
}

func (b *spoolBuffer) writeFile(p []byte) (int, error) {
	if _, err := b.file.Seek(b.size, io.SeekStart); err != nil {
		return 0, err
	}
	n, err := writeOnce(b.file, p)
	b.size += int64(n)
	return n, err
}

func (b *spoolBuffer) writeFileString(s string) (int, error) {
	if _, err := b.file.Seek(b.size, io.SeekStart); err != nil {
		return 0, err
	}
	n, err := b.file.WriteString(s)
	n, err = normalizeWriteResult(n, len(s), err)
	b.size += int64(n)
	return n, err
}

func writeOnce(w io.Writer, p []byte) (int, error) {
	n, err := w.Write(p)
	return normalizeWriteResult(n, len(p), err)
}

func normalizeWriteResult(n, want int, err error) (int, error) {
	if n < 0 || n > want {
		return 0, errors.New("middleware: invalid Write count")
	}
	if n != want && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

type readerOnly struct {
	io.Reader
}
