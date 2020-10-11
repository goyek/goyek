package taskflow

import (
	"context"
	"fmt"
	"io"
	"runtime"
)

// TF TODO.
type TF struct {
	ctx     context.Context
	writer  io.Writer
	failed  bool
	skipped bool
}

// Context TODO.
func (tf *TF) Context() context.Context {
	return tf.ctx
}

// Name TODO.
func Name() string {
	panic("TODO")
}

// Writer TODO.
func (tf *TF) Writer() io.Writer {
	return tf.writer
}

// Logf TODO.
func (tf *TF) Logf(format string, args ...interface{}) {
	fmt.Fprintf(tf.writer, format+"\n", args...)
}

// Errorf TODO.
func (tf *TF) Errorf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.Fail()
}

// Fail TODO.
func (tf *TF) Fail() {
	tf.failed = true
}

// Fatalf TODO.
func (tf *TF) Fatalf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.FailNow()
}

// FailNow TODO.
func (tf *TF) FailNow() {
	tf.Fail()
	runtime.Goexit()
}

// Skipf TODO.
func (tf *TF) Skipf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.SkipNow()
}

// SkipNow TODO.
func (tf *TF) SkipNow() {
	tf.skipped = true
	runtime.Goexit()
}

// Cleanup TODO.
func (tf *TF) Cleanup(func()) {
	panic("TODO")
}

// Helper TODO.
// func (tf *TF) Helper() {
// 	panic("TODO")
// }
