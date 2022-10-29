package goyek

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
)

// Logger is used by TF's logging functions.
type Logger interface {
	Log(w io.Writer, args ...interface{})
	Logf(w io.Writer, format string, args ...interface{})
}

// FmtLogger uses fmt when logging. It only appends new line at the end.
type FmtLogger struct{}

// Log is used internally in order to provide proper prefix.
func (l FmtLogger) Log(w io.Writer, args ...interface{}) {
	fmt.Fprintln(w, args...)
}

// Logf is used internally in order to provide proper prefix.
func (l FmtLogger) Logf(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+"\n", args...)
}

// CodeLineLogger decorates the log with code line information and identation.
type CodeLineLogger struct {
	mu          sync.Mutex
	helperPCs   map[uintptr]struct{} // functions to be skipped when writing file/line info
	helperNames map[string]struct{}  // helperPCs converted to function names
}

// Log is used internally in order to provide proper prefix.
func (l *CodeLineLogger) Log(w io.Writer, args ...interface{}) {
	txt := fmt.Sprint(args...)
	txt = l.decorate(txt)
	io.WriteString(w, txt) //nolint:errcheck,gosec // not checking errors when writing to output
}

// Logf is used internally in order to provide proper prefix.
func (l *CodeLineLogger) Logf(w io.Writer, format string, args ...interface{}) {
	txt := fmt.Sprintf(format, args...)
	txt = l.decorate(txt)
	io.WriteString(w, txt) //nolint:errcheck,gosec // not checking errors when writing to output
}

// Helper marks the calling function as a helper function.
// When printing file and line information, that function will be skipped.
// Helper may be called simultaneously from multiple goroutines.
func (l *CodeLineLogger) Helper() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.helperPCs == nil {
		l.helperPCs = make(map[uintptr]struct{})
	}
	// repeating code from callerName here to save walking a stack frame
	var pc [1]uintptr
	const skip = 3
	n := runtime.Callers(skip, pc[:]) // skip runtime.Callers + CodeLineLogger.Helper + TF.Helper
	if n == 0 {
		panic("zero callers found")
	}
	if _, found := l.helperPCs[pc[0]]; !found {
		l.helperPCs[pc[0]] = struct{}{}
		l.helperNames = nil // map will be recreated next time it is needed
	}
}

// decorate prefixes the string with the file and line of the call site
// and inserts the final newline and indentation spaces for formatting.
func (*CodeLineLogger) decorate(s string) string {
	const skip = 3
	_, file, line, _ := runtime.Caller(skip)
	if file != "" {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
	}
	if line == 0 {
		line = 1
	}
	buf := &strings.Builder{}
	// Every line is indented at least 6 spaces.
	buf.WriteString("      ")
	fmt.Fprintf(buf, "%s:%d: ", file, line)
	lines := strings.Split(s, "\n")
	if l := len(lines); l > 1 && lines[l-1] == "" {
		lines = lines[:l-1]
	}
	for i, line := range lines {
		if i > 0 {
			// Second and subsequent lines are indented an additional 4 spaces.
			buf.WriteString("\n          ")
		}
		buf.WriteString(line)
	}
	buf.WriteByte('\n')
	return buf.String()
}
