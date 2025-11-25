package middleware

import (
	"fmt"
	"time"

	"github.com/goyek/goyek/v3"
)

// ReportFlow is a middleware which reports the flow execution status.
//
// The format is based on the reports provided by the Go test runner.
func ReportFlow(next goyek.Executor) goyek.Executor {
	return func(in goyek.ExecuteInput) error {
		out := in.Output

		from := time.Now()
		if err := next(in); err != nil {
			fmt.Fprintf(out, "%v\t%.3fs\n", err, time.Since(from).Seconds())
			return err
		}
		fmt.Fprintf(out, "ok\t%.3fs\n", time.Since(from).Seconds())
		return nil
	}
}
