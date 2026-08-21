package middleware

import (
	"io"
	"os"
	"strings"
)

const (
	maxBufferedOutputBytes = 1 << 20
	outputSpillPattern     = "goyek-output-*"
)

// spoolBuffer retains small output in memory and spills larger output to a
// temporary file. It is not safe for concurrent use; middleware wraps it with
// goyek.SyncWriter before exposing it to a runner.
type spoolBuffer struct {
	buffer strings.Builder
	file   *os.File
	limit  int
	size   int64
}

func newSpoolBuffer(limit int) *spoolBuffer {
	return &spoolBuffer{limit: limit}
}

func (b *spoolBuffer) Write(p []byte) (int, error) {
	if b.file == nil && len(p) <= b.limit-b.buffer.Len() {
		n, err := b.buffer.Write(p)
		b.size += int64(n)
		return n, err
	}
	if b.file == nil {
		if err := b.spill(); err != nil {
			return 0, err
		}
	}
	n, err := b.file.Write(p)
	b.size += int64(n)
	if n < len(p) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

func (b *spoolBuffer) WriteString(s string) (int, error) {
	if b.file == nil && len(s) <= b.limit-b.buffer.Len() {
		n, err := b.buffer.WriteString(s)
		b.size += int64(n)
		return n, err
	}
	if b.file == nil {
		if err := b.spill(); err != nil {
			return 0, err
		}
	}
	n, err := io.WriteString(b.file, s)
	b.size += int64(n)
	if n < len(s) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

func (b *spoolBuffer) spill() error {
	file, err := os.CreateTemp("", outputSpillPattern)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(file, b.buffer.String()); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return err
	}
	b.buffer.Reset()
	b.file = file
	return nil
}

func (b *spoolBuffer) Len() int64 {
	return b.size
}

// reader returns a reader starting at the beginning of the retained output.
// Its concrete WriterTo method is hidden so a destination ReaderFrom method
// (notably goyek.SyncWriter's) controls the copy and its synchronization unit.
func (b *spoolBuffer) reader() (io.Reader, error) {
	if b.file == nil {
		return struct{ io.Reader }{strings.NewReader(b.buffer.String())}, nil
	}
	if _, err := b.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return struct{ io.Reader }{b.file}, nil
}

func (b *spoolBuffer) WriteTo(output io.Writer) (int64, error) {
	reader, err := b.reader()
	if err != nil {
		return 0, err
	}
	return io.Copy(output, reader)
}

func (b *spoolBuffer) Reset() error {
	b.buffer.Reset()
	b.size = 0
	return b.closeFile()
}

func (b *spoolBuffer) Close() error {
	return b.Reset()
}

func (b *spoolBuffer) closeFile() error {
	if b.file == nil {
		return nil
	}

	name := b.file.Name()
	closeErr := b.file.Close()
	removeErr := os.Remove(name)
	b.file = nil
	if closeErr != nil {
		return closeErr
	}
	return removeErr
}
