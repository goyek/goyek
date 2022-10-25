package middleware_test

import (
	"testing"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func TestDryRun(t *testing.T) {
	called := false
	r := func(i goyek.Input) goyek.Result {
		called = true
		return goyek.Result{Status: goyek.StatusPassed}
	}
	r = middleware.DryRun(r)

	got := r(goyek.Input{})

	if called {
		t.Error("should not call the next runner")
	}
	if got.Status != goyek.StatusNotRun {
		t.Errorf("should have status not run but was: %v", got.Status)
	}
}
