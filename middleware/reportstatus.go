package middleware

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/goyek/goyek/v3"
)

// ReportStatus is a middleware which reports the task run status.
//
// The format is based on the reports provided by the Go test runner.
func ReportStatus(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		out := outputOrDiscard(in.Output)
		in.Output = out

		// report start task
		fmt.Fprintf(out, "===== TASK  %s\n", in.TaskName)
		start := time.Now()

		// run
		res := next(in)

		// report task end
		fmt.Fprintf(out, "----- %s: %s (%.2fs)\n", res.Status, in.TaskName, time.Since(start).Seconds())

		// report panic if happened
		if res.PanicStack != nil {
			var report strings.Builder
			if res.PanicValue != nil {
				fmt.Fprintf(&report, "panic: %v", res.PanicValue)
			} else {
				report.WriteString("panic(nil) or runtime.Goexit() called")
			}
			report.WriteString("\n\n")
			report.Write(res.PanicStack)
			io.WriteString(out, report.String()) //nolint:errcheck // not checking errors when writing to output
		}

		return res
	}
}
