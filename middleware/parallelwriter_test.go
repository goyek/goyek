package middleware

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

func TestParallelWriter(t *testing.T) {
	tests := []struct {
		name  string
		write func(io.Writer) (int, error)
	}{
		{
			name: "Write",
			write: func(w io.Writer) (int, error) {
				return w.Write([]byte("message\n"))
			},
		},
		{
			name: "WriteString",
			write: func(w io.Writer) (int, error) {
				return io.WriteString(w, "message\n")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			writer := newParallelWriter(&output, "task")
			t.Cleanup(func() { _ = writer.Close() })

			n, err := tt.write(writer)
			if err != nil {
				t.Fatalf("write returned error: %v", err)
			}
			if n != len("message\n") {
				t.Fatalf("write returned %d, want %d", n, len("message\n"))
			}
			if got, want := output.String(), "=== NAME  task\nmessage\n"; got != want {
				t.Fatalf("output = %q, want %q", got, want)
			}
		})
	}
}

func TestParallelWriter_shortWrite(t *testing.T) {
	writer := newParallelWriter(&shortWriter{remaining: len("=== NAME  task\n") + 2}, "task")
	t.Cleanup(func() { _ = writer.Close() })

	n, err := writer.Write([]byte("message\n"))

	if n != 2 {
		t.Fatalf("Write returned %d, want 2", n)
	}
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("Write returned error %v, want %v", err, io.ErrShortWrite)
	}
}

func TestParallelWriter_multilineWrite(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")
	t.Cleanup(func() { _ = writer.Close() })

	n, err := io.WriteString(writer, "first\nsecond\n")
	if err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if n != len("first\nsecond\n") {
		t.Fatalf("WriteString returned %d, want %d", n, len("first\nsecond\n"))
	}
	want := "=== NAME  task\nfirst\n=== NAME  task\nsecond\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestParallelWriter_splitLine(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")
	t.Cleanup(func() { _ = writer.Close() })

	if n, err := writer.Write([]byte("hel")); err != nil || n != 3 {
		t.Fatalf("first Write returned %d, %v; want 3, nil", n, err)
	}
	if output.Len() != 0 {
		t.Fatalf("partial line was streamed early: %q", output.String())
	}
	if n, err := writer.WriteString("lo\n"); err != nil || n != 3 {
		t.Fatalf("second WriteString returned %d, %v; want 3, nil", n, err)
	}

	if got, want := output.String(), "=== NAME  task\nhello\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestParallelWriter_splitUTF8Rune(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")
	t.Cleanup(func() { _ = writer.Close() })
	message := []byte("€\n")

	if n, err := writer.Write(message[:1]); err != nil || n != 1 {
		t.Fatalf("first Write returned %d, %v; want 1, nil", n, err)
	}
	if n, err := writer.Write(message[1:]); err != nil || n != len(message)-1 {
		t.Fatalf("second Write returned %d, %v; want %d, nil", n, err, len(message)-1)
	}

	if got, want := output.String(), "=== NAME  task\n€\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestParallelWriter_flushesPartialLine(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")

	if n, err := writer.WriteString("message"); err != nil || n != len("message") {
		t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len("message"))
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if got, want := output.String(), "=== NAME  task\nmessage\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestParallelWriter_Flush(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")

	if n, err := writer.WriteString("first"); err != nil || n != len("first") {
		t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len("first"))
	}
	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
	if n, err := writer.WriteString("second\n"); err != nil || n != len("second\n") {
		t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len("second\n"))
	}
	if got, want := output.String(), "=== NAME  task\nfirst\n=== NAME  task\nsecond\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestParallelWriter_afterClose(t *testing.T) {
	writer := newParallelWriter(io.Discard, "task")
	if err := writer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if n, err := writer.Write([]byte("message")); n != 0 || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Write returned %d, %v; want 0, %v", n, err, io.ErrClosedPipe)
	}
	if n, err := writer.WriteString("message"); n != 0 || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("WriteString returned %d, %v; want 0, %v", n, err, io.ErrClosedPipe)
	}
	if n, err := writer.ReadFrom(strings.NewReader("message")); n != 0 || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("ReadFrom returned %d, %v; want 0, %v", n, err, io.ErrClosedPipe)
	}
	if err := writer.Flush(); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Flush returned %v, want %v", err, io.ErrClosedPipe)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("second Close returned error: %v", err)
	}
}

func TestParallelWriter_truncatesLongLine(t *testing.T) {
	tests := []struct {
		name  string
		write func(io.Writer, string) (int, error)
	}{
		{
			name: "Write",
			write: func(w io.Writer, s string) (int, error) {
				return w.Write([]byte(s))
			},
		},
		{name: "WriteString", write: io.WriteString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			writer := newParallelWriter(&output, "task")
			t.Cleanup(func() { _ = writer.Close() })
			message := strings.Repeat("x", maxBufferedOutputBytes+1) + "ignored\nnext\n"

			n, err := tt.write(writer, message)
			if err != nil {
				t.Fatalf("write returned error: %v", err)
			}
			if n != len(message) {
				t.Fatalf("write returned %d, want %d", n, len(message))
			}

			want := "=== NAME  task\n" + strings.Repeat("x", maxBufferedOutputBytes) +
				parallelTruncatedMarker + "=== NAME  task\nnext\n"
			if got := output.String(); got != want {
				t.Fatalf("output length = %d, want %d", len(got), len(want))
			}
		})
	}
}

func TestParallelWriter_ReadFrom(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")
	t.Cleanup(func() { _ = writer.Close() })
	message := "first\nsecond\n"

	n, err := writer.ReadFrom(strings.NewReader(message))
	if err != nil {
		t.Fatalf("ReadFrom returned error: %v", err)
	}
	if n != int64(len(message)) {
		t.Fatalf("ReadFrom returned %d, want %d", n, len(message))
	}
	want := "=== NAME  task\nfirst\n=== NAME  task\nsecond\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestParallelWriter_truncationPreservesUTF8(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")
	t.Cleanup(func() { _ = writer.Close() })
	message := strings.Repeat("x", maxBufferedOutputBytes-1) + "€more\n"

	if n, err := writer.WriteString(message); err != nil || n != len(message) {
		t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
	}

	want := "=== NAME  task\n" + strings.Repeat("x", maxBufferedOutputBytes-1) + parallelTruncatedMarker
	if got := output.String(); got != want {
		t.Fatalf("output length = %d, want %d", len(got), len(want))
	}
	if !utf8.ValidString(output.String()) {
		t.Fatal("truncated output is not valid UTF-8")
	}
}

func TestParallelWriter_truncationAcrossWritesPreservesUTF8(t *testing.T) {
	var output bytes.Buffer
	writer := newParallelWriter(&output, "task")
	t.Cleanup(func() { _ = writer.Close() })
	runeBytes := []byte("€")
	first := append([]byte(strings.Repeat("x", maxBufferedOutputBytes-2)), runeBytes[:2]...)

	if n, err := writer.Write(first); err != nil || n != len(first) {
		t.Fatalf("first Write returned %d, %v; want %d, nil", n, err, len(first))
	}
	if n, err := writer.Write(append(runeBytes[2:], []byte("discarded")...)); err != nil || n != len("€")-2+len("discarded") {
		t.Fatalf("second Write returned %d, %v; want %d, nil", n, err, len("€")-2+len("discarded"))
	}

	want := "=== NAME  task\n" + strings.Repeat("x", maxBufferedOutputBytes-2) + parallelTruncatedMarker
	if got := output.String(); got != want {
		t.Fatalf("output length = %d, want %d", len(got), len(want))
	}
	if !utf8.ValidString(output.String()) {
		t.Fatal("truncated output is not valid UTF-8")
	}
}

func TestTrimIncompleteUTF8_shortInput(t *testing.T) {
	input := []byte{0xe2}
	if got := trimIncompleteUTF8(input); len(got) != 0 {
		t.Fatalf("trimIncompleteUTF8 returned %x, want empty output", got)
	}
}

func TestParallelWriter_invalidOutputCount(t *testing.T) {
	tests := []struct {
		name   string
		output io.Writer
		wantN  int
	}{
		{
			name: "negative",
			output: writerFunc(func([]byte) (int, error) {
				return -1, nil
			}),
		},
		{
			name: "too large",
			output: writerFunc(func(p []byte) (int, error) {
				return len(p) + 1, nil
			}),
			wantN: len("message\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := newParallelWriter(tt.output, "task")
			n, err := writer.WriteString("message\n")
			if n != tt.wantN || !errors.Is(err, io.ErrShortWrite) {
				t.Fatalf("WriteString returned %d, %v; want %d, %v", n, err, tt.wantN, io.ErrShortWrite)
			}
		})
	}
}

func TestParallelWriter_truncationStateSurvivesWriteError(t *testing.T) {
	tests := []struct {
		name  string
		write func(io.Writer, string) (int, error)
	}{
		{
			name: "Write",
			write: func(w io.Writer, s string) (int, error) {
				return w.Write([]byte(s))
			},
		},
		{name: "WriteString", write: io.WriteString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &fullWriteErrorWriter{err: errors.New("write failed")}
			writer := newParallelWriter(output, "task")
			t.Cleanup(func() { _ = writer.Close() })
			message := strings.Repeat("x", maxBufferedOutputBytes+1) + "\nnext\n"

			n, err := tt.write(writer, message)
			if n != maxBufferedOutputBytes {
				t.Fatalf("first write returned %d, want %d", n, maxBufferedOutputBytes)
			}
			if !errors.Is(err, output.err) {
				t.Fatalf("first write returned error %v, want %v", err, output.err)
			}
			if retryN, retryErr := tt.write(writer, message[n:]); retryErr != nil || retryN != len(message)-n {
				t.Fatalf("retry returned %d, %v; want %d, nil", retryN, retryErr, len(message)-n)
			}

			want := "=== NAME  task\n" + strings.Repeat("x", maxBufferedOutputBytes) +
				parallelTruncatedMarker + "=== NAME  task\nnext\n"
			if got := output.String(); got != want {
				t.Fatalf("output length = %d, want %d", len(got), len(want))
			}
		})
	}
}

func TestParallelWriter_partialTruncationWriteErrorIsSticky(t *testing.T) {
	writes := []struct {
		name  string
		write func(io.Writer, string) (int, error)
	}{
		{
			name: "Write",
			write: func(w io.Writer, s string) (int, error) {
				return w.Write([]byte(s))
			},
		},
		{name: "WriteString", write: io.WriteString},
	}
	partialWrites := []struct {
		name           string
		outputBytes    int
		wantInputBytes int
	}{
		{name: "zero", outputBytes: 0, wantInputBytes: 0},
		{name: "short", outputBytes: len("=== NAME  task\n") + 2, wantInputBytes: 2},
	}

	for _, write := range writes {
		for _, partial := range partialWrites {
			t.Run(write.name+"/"+partial.name, func(t *testing.T) {
				writeErr := errors.New("write failed")
				output := &partialWriteErrorWriter{n: partial.outputBytes, err: writeErr}
				writer := newParallelWriter(output, "task")
				t.Cleanup(func() { _ = writer.Close() })
				message := strings.Repeat("x", maxBufferedOutputBytes+1) + "\n"

				n, err := write.write(writer, message)
				if n != partial.wantInputBytes {
					t.Fatalf("first write returned %d, want %d", n, partial.wantInputBytes)
				}
				if !errors.Is(err, writeErr) {
					t.Fatalf("first write returned error %v, want %v", err, writeErr)
				}
				retryN, retryErr := write.write(writer, message[n:])
				if retryN != 0 || !errors.Is(retryErr, writeErr) {
					t.Fatalf("retry returned %d, %v; want 0, %v", retryN, retryErr, writeErr)
				}
				if output.calls != 1 {
					t.Fatalf("underlying Write calls = %d, want 1", output.calls)
				}
				if readN, readErr := writer.ReadFrom(strings.NewReader("ignored")); readN != 0 || !errors.Is(readErr, writeErr) {
					t.Fatalf("ReadFrom after write error returned %d, %v; want 0, %v", readN, readErr, writeErr)
				}
				if flushErr := writer.Flush(); !errors.Is(flushErr, writeErr) {
					t.Fatalf("Flush after write error returned %v, want %v", flushErr, writeErr)
				}
			})
		}
	}
}

func TestParallelWriter_sanitizesTaskName(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		wantName string
	}{
		{name: "line breaks", taskName: "line\r\nbreak", wantName: `line\r\nbreak`},
		{name: "literal escapes", taskName: `line\r\nbreak`, wantName: `line\\r\\nbreak`},
		{name: "other controls", taskName: "tab\tzero\x00", wantName: `tab\tzero\x00`},
		{name: "graphic Unicode", taskName: "ą😊", wantName: "ą😊"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			writer := newParallelWriter(&output, tt.taskName)
			t.Cleanup(func() { _ = writer.Close() })

			if _, err := writer.WriteString("message\n"); err != nil {
				t.Fatalf("WriteString returned error: %v", err)
			}

			want := "=== NAME  " + tt.wantName + "\nmessage\n"
			if got := output.String(); got != want {
				t.Fatalf("output = %q, want %q", got, want)
			}
		})
	}
}

func TestParallelWriter_recordsUseOneWrite(t *testing.T) {
	output := &recordingWriter{}
	first := newParallelWriter(output, "first")
	second := newParallelWriter(output, "second")
	start := make(chan struct{})
	var wait sync.WaitGroup
	for _, writer := range []*parallelWriter{first, second} {
		writer := writer
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, _ = writer.WriteString("message\n")
		}()
	}
	close(start)
	wait.Wait()

	if len(output.records) != 2 {
		t.Fatalf("underlying Write calls = %d, want 2: %q", len(output.records), output.records)
	}
	for _, record := range output.records {
		if record != "=== NAME  first\nmessage\n" && record != "=== NAME  second\nmessage\n" {
			t.Fatalf("split or malformed record: %q", record)
		}
	}
}

func TestParallelWriter_manyLinesHasBoundedAllocations(t *testing.T) {
	message := strings.Repeat("x\n", 4096)
	allocations := testing.AllocsPerRun(5, func() {
		output := &countWriter{}
		writer := newParallelWriter(output, "task")
		if n, err := writer.WriteString(message); err != nil || n != len(message) {
			t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})
	if allocations > 20 {
		t.Fatalf("allocations per run = %.0f, want at most 20", allocations)
	}
}

type recordingWriter struct {
	mu      sync.Mutex
	records []string
}

type fullWriteErrorWriter struct {
	bytes.Buffer
	err   error
	calls int
}

type partialWriteErrorWriter struct {
	n     int
	err   error
	calls int
}

func (w *partialWriteErrorWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.n > len(p) {
		return len(p), w.err
	}
	return w.n, w.err
}

func (w *fullWriteErrorWriter) Write(p []byte) (int, error) {
	w.calls++
	_, _ = w.Buffer.Write(p)
	if w.calls == 1 {
		return len(p), w.err
	}
	return len(p), nil
}

func (w *recordingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.records = append(w.records, string(p))
	return len(p), nil
}

type shortWriter struct {
	remaining int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if w.remaining >= len(p) {
		w.remaining -= len(p)
		return len(p), nil
	}
	n := w.remaining
	w.remaining = 0
	return n, nil
}
