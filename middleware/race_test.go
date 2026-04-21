package middleware_test

import (
	"io"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/middleware"
)

func TestBufferParallel_Race(_ *testing.T) {
	runner := middleware.BufferParallel(func(in goyek.Input) goyek.Result {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					io.WriteString(in.Output, "a") //nolint:errcheck // not checking errors when writing to output
				}
			}()
		}
		wg.Wait()
		return goyek.Result{Status: goyek.StatusPassed}
	})

	in := goyek.Input{
		Parallel: true,
		Output:   io.Discard,
	}
	runner(in)
}

func TestSilentNonFailed_Race(_ *testing.T) {
	runner := middleware.SilentNonFailed(func(in goyek.Input) goyek.Result {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					io.WriteString(in.Output, "a") //nolint:errcheck // not checking errors when writing to output
				}
			}()
		}
		wg.Wait()
		return goyek.Result{Status: goyek.StatusFailed}
	})

	in := goyek.Input{
		Output: io.Discard,
	}
	runner(in)
}
