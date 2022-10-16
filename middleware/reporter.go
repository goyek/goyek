package middleware

import (
	"fmt"
	"io"
	"time"

	"github.com/goyek/goyek/v2"
)

// Reporter banana.
func Reporter(next goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		// report start task
		fmt.Fprintf(in.Output, "===== TASK  %s\n", in.TaskName)
		start := time.Now()

		// run
		res := next(in)

		// report task end
		status := "PASS"
		switch res.Status {
		case goyek.StatusFailed:
			status = "FAIL"
		case goyek.StatusNotRun, goyek.StatusSkipped:
			status = "SKIP"
		}
		fmt.Fprintf(in.Output, "----- %s: %s (%.2fs)\n", status, in.TaskName, time.Since(start).Seconds())

		// report panic if happened
		if res.PanicStack != nil {
			if res.PanicValue != nil {
				io.WriteString(in.Output, fmt.Sprintf("panic: %v", res.PanicValue)) //nolint:errcheck,gosec // not checking errors when writing to output
			} else {
				io.WriteString(in.Output, "panic(nil) or runtime.Goexit() called") //nolint:errcheck,gosec // not checking errors when writing to output
			}
			io.WriteString(in.Output, "\n\n") //nolint:errcheck,gosec // not checking errors when writing to output
			in.Output.Write(res.PanicStack)   //nolint:errcheck,gosec // not checking errors when writing to output
		}

		return res
	}
}
