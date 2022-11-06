package middleware_test

import (
	"strings"
	"testing"
	"time"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/goyek/v2/middleware"
)

func TestReportLongRun(t *testing.T) {
	taskName := "my-task"
	sb := &strings.Builder{}
	r := goyek.NewRunner(func(a *goyek.A) { time.Sleep(time.Second) })
	r = middleware.ReportLongRun(10 * time.Millisecond)(r)

	r(goyek.Input{TaskName: taskName, Output: sb})

	if !strings.Contains(sb.String(), "***** LONG: "+taskName+" (") {
		t.Errorf("got: %q; but should long running report", sb.String())
	}
}
