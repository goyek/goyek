package goyek

import (
	"fmt"
	"io"
	"time"
)

func reporter(next runner) runner {
	return func(in input) result {
		// report start task
		fmt.Fprintf(in.Output, "===== TASK  %s\n", in.TaskName)
		start := time.Now()

		// run
		res := next(in)

		// report task end
		status := "PASS"
		switch res.status {
		case statusFailed, statusPanicked:
			status = "FAIL"
		case statusNotRun, statusSkipped:
			status = "SKIP"
		}
		fmt.Fprintf(in.Output, "----- %s: %s (%.2fs)\n", status, in.TaskName, time.Since(start).Seconds())

		// report panic if happened
		if res.status == statusPanicked {
			if res.panicValue != nil {
				io.WriteString(in.Output, fmt.Sprintf("panic: %v", res.panicValue)) //nolint:errcheck,gosec // not checking errors when writing to output
			} else {
				io.WriteString(in.Output, "panic(nil) or runtime.Goexit() called") //nolint:errcheck,gosec // not checking errors when writing to output
			}
			io.WriteString(in.Output, "\n\n") //nolint:errcheck,gosec // not checking errors when writing to output
			in.Output.Write(res.panicStack)   //nolint:errcheck,gosec // not checking errors when writing to output
		}

		return res
	}
}
