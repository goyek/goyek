package middleware

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/internal"
)

// ReportLongRun is a middleware which reports the task when it is long running.
// It may pass a synchronized wrapper around [goyek.Input.Output] to the next
// runner. The next runner must not rely on the writer's concrete type or
// optional interfaces, and must use the writer it receives instead of a
// previously retained reference.
func ReportLongRun(d time.Duration) func(next goyek.Runner) goyek.Runner {
	return func(next goyek.Runner) goyek.Runner {
		return func(in goyek.Input) goyek.Result {
			out := in.Output
			if out == nil {
				out = io.Discard
			}
			in.Output = internal.SyncWriter(out)

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

			defer func() {
				close(done)
				wg.Wait()
			}()
			return next(in)
		}
	}
}
