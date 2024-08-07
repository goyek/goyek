package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/goyek/goyek/v2"
)

// ReportStatus is a middleware which reports the flow execution status.
//
// The format is based on the reports provided by the Go test runner.
func ReportFlow(next goyek.Executor) goyek.Executor {
	return func(in goyek.ExecuteInput) error {
		out := in.Output

		from := time.Now()
		err := next(in)
		if _, ok := err.(*goyek.FailError); ok {
			fmt.Fprintf(out, "%v\t%.3fs\n", err, time.Since(from).Seconds())
			return err
		}
		if err == context.Canceled || err == context.DeadlineExceeded {
			fmt.Fprintf(out, "%v\t%.3fs\n", err, time.Since(from).Seconds())
			return err
		}
		if err != nil {
			fmt.Fprintln(out, err.Error())
			return err
		}
		fmt.Fprintf(out, "ok\t%.3fs\n", time.Since(from).Seconds())
		return nil
	}
}
