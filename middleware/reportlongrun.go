package middleware

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/goyek/goyek/v3"
)

// ReportLongRun is a middleware which reports the task when it is long running.
// It can write to [goyek.Input.Output] concurrently with the next runner, so a
// non-nil output must be safe for concurrent use. Use [goyek.SyncWriter] to
// adapt a writer that does not provide its own synchronization. A nil output
// is replaced with [io.Discard].
func ReportLongRun(d time.Duration) func(next goyek.Runner) goyek.Runner {
	return func(next goyek.Runner) goyek.Runner {
		return func(in goyek.Input) goyek.Result {
			out := in.Output
			if out == nil {
				out = io.Discard
			}
			in.Output = out

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
						fmt.Fprintf(out, "***** LONG: %s (%.2fs)\n", task, time.Since(start).Seconds())
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
