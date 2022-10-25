package middleware_test

import (
	"testing"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func TestSkip_skipped(t *testing.T) {
	called := false
	r := func(i goyek.Input) goyek.Result {
		called = true
		return goyek.Result{Status: goyek.StatusPassed}
	}
	r = middleware.Skip([]string{"task"})(r)

	got := r(goyek.Input{TaskName: "task"})

	if called {
		t.Error("should not call the next runner")
	}
	if got.Status != goyek.StatusNotRun {
		t.Errorf("should have status not run but was: %v", got.Status)
	}
}

func TestSkip_not_skipped(t *testing.T) {
	called := false
	r := func(i goyek.Input) goyek.Result {
		called = true
		return goyek.Result{Status: goyek.StatusPassed}
	}
	r = middleware.Skip([]string{"other"})(r)

	got := r(goyek.Input{TaskName: "task"})

	if !called {
		t.Error("should call the next runner")
	}
	if got.Status != goyek.StatusPassed {
		t.Errorf("should have status from the task run but was: %v", got.Status)
	}
}
