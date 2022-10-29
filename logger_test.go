package goyek_test

import (
	"context"
	"strings"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestCodeLineLogger(t *testing.T) {
	flow := &goyek.Flow{}
	out := &strings.Builder{}
	flow.SetOutput(out)
	loggerSpy := &goyek.CodeLineLogger{}
	flow.SetLogger(loggerSpy)
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(tf *goyek.TF) {
			tf.Log("message")
			helperFn(tf)
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "      logger_test.go:20: message", "should contain code line info")
	assertContains(t, out, "      logger_test.go:21: message from helper", "should respect tf.Helper()")
}

func helperFn(tf *goyek.TF) {
	tf.Helper()
	tf.Log("message from helper")
}
