package goyek_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestFlow_Concurrent(t *testing.T) {
	f := &goyek.Flow{}
	ctx := context.Background()

	var wg sync.WaitGroup
	const iterations = 100

	// Concurrent registration
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			f.Define(goyek.Task{
				Name: fmt.Sprintf("task-%d", i),
				Action: func(a *goyek.A) {
					a.Log("running")
				},
			})
		}
	}()

	// Concurrent modification of defined tasks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			taskName := fmt.Sprintf("mod-task-%d", i)
			task := f.Define(goyek.Task{Name: taskName})
			task.SetUsage("some usage")
			task.SetAction(func(a *goyek.A) {})
			task.SetName(taskName + "-renamed")
		}
	}()

	// Concurrent execution
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = f.Execute(ctx, []string{"task-0"})
		}
	}()

	// Concurrent flow property changes
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			f.SetOutput(nil)
			f.SetLogger(nil)
			f.SetUsage(nil)
		}
	}()

	wg.Wait()
}
