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
		Action: func(a *goyek.A) {
			a.Log("message")
			helperFn(a)
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "      logger_test.go:20: message", "should contain code line info")
	assertContains(t, out, "      logger_test.go:21: message from helper", "should respect a.Helper()")
}

func TestCodeLineLogger_helper_in_action(t *testing.T) {
	flow := &goyek.Flow{}
	out := &strings.Builder{}
	flow.SetOutput(out)
	loggerSpy := &goyek.CodeLineLogger{}
	flow.SetLogger(loggerSpy)
	flow.Define(goyek.Task{
		Name: "task",
		Action: func(a *goyek.A) {
			a.Helper()
			a.Log("message")
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "      logger_test.go:41: message", "should contain code line info")
}

func helperFn(a *goyek.A) {
	a.Helper()
	a.Log("message from helper")
}
