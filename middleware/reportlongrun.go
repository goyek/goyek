package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/goyek/goyek/v3"
)

// ReportLongRun is a middleware which reports the task when it is long running.
func ReportLongRun(d time.Duration) func(next goyek.Runner) goyek.Runner {
	return func(next goyek.Runner) goyek.Runner {
		return func(in goyek.Input) goyek.Result {
			start := time.Now()
			task := in.TaskName
			done := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				t := time.NewTicker(d)
				defer t.Stop()
				for {
					select {
					case <-done:
						return
					case <-t.C:
						fmt.Fprintf(in.Output, "***** LONG: %s (%.2fs)\n", task, time.Since(start).Seconds())
					}
				}
			}()

			// run
			res := next(in)
			close(done)
			wg.Wait()
			return res
		}
	}
}
