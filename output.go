package goyek

import (
	"fmt"
	"io"
	"sync"
)

// Output contains the writers to communicate results.
type Output struct {
	// Primary output is for information that could be processed by piped commands.
	Primary io.Writer
	// Message output is for any status information. Goyek only prints to this writer.
	Message io.Writer
}

func (out Output) WriteMessage(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(out.Message, format+"\n", a...)
}

// bufferedOutput stores all written data in memory.
type bufferedOutput struct {
	mutex   sync.Mutex
	entries []*bufferedEntry
}

type bufferedEntry struct {
	primary bool
	data    []byte
}

// Output returns an Output instance with writers that store written data in this buffer.
func (buffer *bufferedOutput) Output() Output {
	return Output{
		Primary: &bufferedWriter{buffer: buffer, primary: true},
		Message: &bufferedWriter{buffer: buffer, primary: false},
	}
}

// WriteTo will reproduce the buffered information into the provided output.
// The buffer keeps track when which writer was used previously and it
// will reproduce the same sequence to the provided output.
func (buffer *bufferedOutput) WriteTo(other Output) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	currentEntries := buffer.entries
	for _, entry := range currentEntries {
		w := other.Message
		if entry.primary {
			w = other.Primary
		}
		_, _ = w.Write(entry.data)
	}
}

func (buffer *bufferedOutput) add(entry *bufferedEntry) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	buffer.entries = append(buffer.entries, entry)
}

type bufferedWriter struct {
	buffer  *bufferedOutput
	primary bool
}

// Write copies the provided data into the backing buffer.
func (w bufferedWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	entry := &bufferedEntry{
		primary: w.primary,
		data:    make([]byte, n),
	}
	copy(entry.data, p)
	w.buffer.add(entry)
	return
}
