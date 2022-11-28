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
			a.Cleanup(func() {
				a.Log("cleanup")
			})
		},
	})

	_ = flow.Execute(context.Background(), []string{"task"})

	assertContains(t, out, "      logger_test.go:20: message", "should contain code line info")
	assertContains(t, out, "      logger_test.go:21: message from helper", "should respect a.Helper()")
	assertContains(t, out, "      logger_test.go:23: cleanup", "should respect a.Cleanup()")
}

func TestCodeLineLoggerHelperInAction(t *testing.T) {
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

	assertContains(t, out, "      logger_test.go:45: message", "should contain code line info")
}

func helperFn(a *goyek.A) {
	a.Helper()
	a.Log("message from helper")
}
