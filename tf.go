package taskflow

import (
	"context"
	"fmt"
	"io"
	"runtime"
)

type TF struct {
	ctx     context.Context
	name    string
	writer  io.Writer
	failed  bool
	skipped bool
}

func (tf *TF) Context() context.Context {
	return tf.ctx
}

func (tf *TF) Name() string {
	return tf.name
}

func (tf *TF) Writer() io.Writer {
	return tf.writer
}

func (tf *TF) Logf(format string, args ...interface{}) {
	fmt.Fprintf(tf.writer, format+"\n", args...)
}

func (tf *TF) Errorf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.Fail()
}

func (tf *TF) Failed() bool {
	return tf.failed
}

func (tf *TF) Fail() {
	tf.failed = true
}

func (tf *TF) Fatalf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.FailNow()
}

func (tf *TF) FailNow() {
	tf.Fail()
	runtime.Goexit()
}

func (tf *TF) Skipped() bool {
	return tf.skipped
}

func (tf *TF) Skipf(format string, args ...interface{}) {
	tf.Logf(format, args...)
	tf.SkipNow()
}

func (tf *TF) SkipNow() {
	tf.skipped = true
	runtime.Goexit()
}
