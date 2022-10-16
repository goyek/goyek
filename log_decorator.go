package goyek

import (
	"fmt"
	"runtime"
	"strings"
)

// LogDecorator is by TF's logging functions to decorate the text.
type LogDecorator interface {
	Decorate(string) string
}

// LogDecoratorFunc implements LogDecorator.
type LogDecoratorFunc func(string) string

// Decorate uses the function to decorate the string.
func (fn LogDecoratorFunc) Decorate(s string) string {
	return fn(s)
}

// CodeLineLogDecorator decorates the log with code line information.
type CodeLineLogDecorator struct{}

// Decorate prefixes the string with the file and line of the call site
// and inserts the final newline and indentation spaces for formatting.
func (*CodeLineLogDecorator) Decorate(s string) string {
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
