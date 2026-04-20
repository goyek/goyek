package middleware_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/goyek/v3/middleware"
)

func TestBufferParallel_Race(_ *testing.T) {
	out := io.Discard
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Use(middleware.BufferParallel)

	flow.Define(goyek.Task{
		Name:     "race",
		Parallel: true,
		Action: func(a *goyek.A) {
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						a.Log("some log message")
					}
				}()
			}
			wg.Wait()
		},
	})

	_ = flow.Execute(context.Background(), []string{"race"})
}

func TestSilentNonFailed_Race(_ *testing.T) {
	out := io.Discard
	flow := &goyek.Flow{}
	flow.SetOutput(out)
	flow.Use(middleware.SilentNonFailed)

	flow.Define(goyek.Task{
		Name: "race",
		Action: func(a *goyek.A) {
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						a.Log("some log message")
					}
				}()
			}
			wg.Wait()
		},
	})

	_ = flow.Execute(context.Background(), []string{"race"})
}
