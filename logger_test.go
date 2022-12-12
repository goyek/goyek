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

	if want, got := "      logger_test.go:20: message", out.String(); !strings.Contains(got, want) {
		t.Errorf("should contain code line info\ngot:%v\nwant substr:%v", got, want)
	}
	if want, got := "      logger_test.go:21: message from helper", out.String(); !strings.Contains(got, want) {
		t.Errorf("should respect a.Helper()\ngot:%v\nwant substr:%v", got, want)
	}
	if want, got := "      logger_test.go:23: cleanup", out.String(); !strings.Contains(got, want) {
		t.Errorf("should respect a.Cleanup()\ngot:%v\nwant substr:%v", got, want)
	}
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

	if want, got := "      logger_test.go:51: message", out.String(); !strings.Contains(got, want) {
		t.Errorf("should contain code line info\ngot:%v\nwant substr:%v", got, want)
	}
}

func helperFn(a *goyek.A) {
	a.Helper()
	a.Log("message from helper")
}
