package middleware

import (
	"fmt"
	"io"
	"time"

	"github.com/goyek/goyek/v2"
)

// ReportStatus is a middleware which reports the task run status.
//
// The format is based on the reports provided by the Go test runner.
func ReportStatus(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		// report start task
		fmt.Fprintf(in.Output, "===== TASK  %s\n", in.TaskName)
		start := time.Now()

		// run
		res := next(in)

		// report task end
		fmt.Fprintf(in.Output, "----- %s: %s (%.2fs)\n", res.Status, in.TaskName, time.Since(start).Seconds())

		// report panic if happened
		if res.PanicStack != nil {
			if res.PanicValue != nil {
				io.WriteString(in.Output, fmt.Sprintf("panic: %v", res.PanicValue)) //nolint:errcheck // not checking errors when writing to output
			} else {
				io.WriteString(in.Output, "panic(nil) or runtime.Goexit() called") //nolint:errcheck // not checking errors when writing to output
			}
			io.WriteString(in.Output, "\n\n") //nolint:errcheck // not checking errors when writing to output
			in.Output.Write(res.PanicStack)   //nolint:errcheck // not checking errors when writing to output
		}

		return res
	}
}
