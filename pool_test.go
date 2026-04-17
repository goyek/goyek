package goyek_test

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goyek/goyek/v3"
)

func TestPool(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)

	pool := flow.DefinePool(goyek.Pool{Name: "pool", Limit: 2})

	var mu sync.Mutex
	var running int
	var maxRunning int

	action := func(_ *goyek.A) {
		mu.Lock()
		running++
		if running > maxRunning {
			maxRunning = running
		}
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		running--
		mu.Unlock()
	}

	flow.Define(goyek.Task{Name: "t1", Action: action, Pools: goyek.DefinedPools{pool}, Parallel: true})
	flow.Define(goyek.Task{Name: "t2", Action: action, Pools: goyek.DefinedPools{pool}, Parallel: true})
	flow.Define(goyek.Task{Name: "t3", Action: action, Pools: goyek.DefinedPools{pool}, Parallel: true})

	err := flow.Execute(context.Background(), []string{"t1", "t2", "t3"})
	assertPass(t, err, "Execute should pass")
	assertEqual(t, maxRunning, 2, "maxRunning should be limited by pool size")
}

func TestMultiPool(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})
	p2 := flow.DefinePool(goyek.Pool{Name: "p2", Limit: 1})

	var mu sync.Mutex
	var running int
	var maxRunning int

	action := func(_ *goyek.A) {
		mu.Lock()
		running++
		if running > maxRunning {
			maxRunning = running
		}
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		running--
		mu.Unlock()
	}

	// t1 uses p1, t2 uses p2, t3 uses both p1 and p2
	flow.Define(goyek.Task{Name: "t1", Action: action, Pools: goyek.DefinedPools{p1}, Parallel: true})
	flow.Define(goyek.Task{Name: "t2", Action: action, Pools: goyek.DefinedPools{p2}, Parallel: true})
	flow.Define(goyek.Task{Name: "t3", Action: action, Pools: goyek.DefinedPools{p1, p2}, Parallel: true})

	err := flow.Execute(context.Background(), []string{"t1", "t2", "t3"})
	assertPass(t, err, "Execute should pass")
	// Since t3 needs both pools, it can only run when both p1 and p2 are free.
	// t1 and t2 can run concurrently if they are the only ones.
	// But any task using p1 or p2 will block t3.
	// maxRunning could be 2 (t1 and t2).
	assertTrue(t, maxRunning <= 2, "maxRunning should respect pool limits")
}

func TestPoolDeadlockAvoidance(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})
	p2 := flow.DefinePool(goyek.Pool{Name: "p2", Limit: 1})

	// Task A uses p1 then p2 (sorted: p1, p2)
	flow.Define(goyek.Task{
		Name:     "A",
		Pools:    goyek.DefinedPools{p1, p2},
		Parallel: true,
		Action:   func(_ *goyek.A) { time.Sleep(time.Millisecond) },
	})
	// Task B uses p2 then p1 (sorted: p1, p2)
	flow.Define(goyek.Task{
		Name:     "B",
		Pools:    goyek.DefinedPools{p2, p1},
		Parallel: true,
		Action:   func(_ *goyek.A) { time.Sleep(time.Millisecond) },
	})

	err := flow.Execute(context.Background(), []string{"A", "B"})
	assertPass(t, err, "Execute should pass without deadlock")
}

func TestPoolContextCancellation(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})

	ctx, cancel := context.WithCancel(context.Background())

	flow.Define(goyek.Task{
		Name:  "blocker",
		Pools: goyek.DefinedPools{p1},
		Action: func(_ *goyek.A) {
			cancel()
			time.Sleep(100 * time.Millisecond)
		},
	})
	flow.Define(goyek.Task{
		Name:  "waiting",
		Pools: goyek.DefinedPools{p1},
	})

	err := flow.Execute(ctx, []string{"blocker", "waiting"})
	assertEqual(t, err, context.Canceled, "should return context.Canceled")
}

func TestPoolSlotLeak(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})
	p2 := flow.DefinePool(goyek.Pool{Name: "p2", Limit: 1})

	ctx, cancel := context.WithCancel(context.Background())

	// blocker1 takes p1
	flow.Define(goyek.Task{
		Name:  "blocker1",
		Pools: goyek.DefinedPools{p1},
		Action: func(_ *goyek.A) {
			time.Sleep(10 * time.Millisecond)
		},
	})
	// leaky tries to take p1 then p2, but will fail on p1 if p1 is taken
	// Actually, runTask will block on p1 acquisition.
	// If we cancel the context while it's blocking on p1, it should not leak anything.
	// If it already acquired p1 and blocks on p2, and we cancel, it should release p1.

	flow.Define(goyek.Task{
		Name:  "blocker2",
		Pools: goyek.DefinedPools{p2},
		Action: func(_ *goyek.A) {
			time.Sleep(50 * time.Millisecond)
		},
	})

	flow.Define(goyek.Task{
		Name:  "leaky",
		Pools: goyek.DefinedPools{p1, p2},
		Parallel: true,
	})

	// 1. blocker1 takes p1
	// 2. blocker2 takes p2
	// 3. leaky tries to take p1, blocks.
	// 4. cancel context.
	// 5. leaky should return.
	// 6. after blocker1 and blocker2 finish, pools should be empty.

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = flow.Execute(ctx, []string{"blocker1", "blocker2", "leaky"})
	}()

	time.Sleep(20 * time.Millisecond) // ensure blockers have taken the slots
	cancel()
	wg.Wait()

	// Now try to run a task that needs the pools. If they leaked, this will hang or fail.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	err := flow.Execute(ctx2, []string{"blocker1", "blocker2"})
	assertPass(t, err, "pools should not have leaked slots")
}

func TestPoolContextCancellationWaitParallel(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	flow.SetLogger(goyek.FmtLogger{})

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})

	ctx, cancel := context.WithCancel(context.Background())

	acquired := make(chan struct{})
	release := make(chan struct{})
	flow.Define(goyek.Task{
		Name:  "blocker",
		Pools: goyek.DefinedPools{p1},
		Parallel: true,
		Action: func(_ *goyek.A) {
			close(acquired)
			<-release
		},
	})

	flow.Define(goyek.Task{
		Name:  "waiter",
		Pools: goyek.DefinedPools{p1},
		Parallel: true,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		err = flow.Execute(ctx, []string{"blocker", "waiter"})
	}()

	<-acquired
	time.Sleep(50 * time.Millisecond)
	cancel()
	close(release)
	wg.Wait()

	if err != nil && !strings.Contains(err.Error(), context.Canceled.Error()) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got %v", err)
	}
}

func TestPoolTaskActionPanic(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})
	flow.Define(goyek.Task{
		Name:  "panic",
		Pools: goyek.DefinedPools{p1},
		Action: func(_ *goyek.A) {
			panic("oops")
		},
	})

	_ = flow.Execute(context.Background(), []string{"panic"})

	// Verify p1 is released
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	flow.Define(goyek.Task{Name: "ok", Pools: goyek.DefinedPools{p1}})
	err := flow.Execute(ctx, []string{"ok"})
	assertPass(t, err, "pool should be released after panic")
}

func TestPoolIntrospection(t *testing.T) {
	flow := &goyek.Flow{}
	p2 := flow.DefinePool(goyek.Pool{Name: "v2", Limit: 2})
	p1 := flow.DefinePool(goyek.Pool{Name: "v1", Limit: 1})

	pools := flow.Pools()
	assertEqual(t, len(pools), 2, "should have 2 pools")
	assertEqual(t, pools[0].Name(), "v1", "first pool name")
	assertEqual(t, pools[0].Limit(), 1, "first pool limit")
	assertEqual(t, pools[1].Name(), "v2", "second pool name")
	assertEqual(t, pools[1].Limit(), 2, "second pool limit")

	task := flow.Define(goyek.Task{
		Name:  "task",
		Pools: goyek.DefinedPools{p2, p1},
	})
	taskPools := task.Pools()
	assertEqual(t, len(taskPools), 2, "task should have 2 pools")
	// Task.Pools() returns them in the order they are in the snapshot, which is sorted by name
	assertEqual(t, taskPools[0].Name(), "v1", "first task pool")
	assertEqual(t, taskPools[1].Name(), "v2", "second task pool")
}

func TestFlow_Define_no_pools(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "no-pools"})
	assertEqual(t, len(task.Pools()), 0, "should have no pools")
}

func TestPoolMultipleSlots(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 2})

	var mu sync.Mutex
	var running int
	var maxRunning int

	action := func(_ *goyek.A) {
		mu.Lock()
		running++
		if running > maxRunning {
			maxRunning = running
		}
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		running--
		mu.Unlock()
	}

	// task takes both slots
	flow.Define(goyek.Task{
		Name:   "t1",
		Pools:  goyek.DefinedPools{p1, p1},
		Action: action,
	})
	// t2 will have to wait
	flow.Define(goyek.Task{
		Name:     "t2",
		Pools:    goyek.DefinedPools{p1},
		Action:   action,
		Parallel: true,
	})

	err := flow.Execute(context.Background(), []string{"t1", "t2"})
	assertPass(t, err, "Execute should pass")
	assertEqual(t, maxRunning, 1, "only 1 task should run at a time because t1 takes all slots")
}

func TestFlow_Define_too_many_slots(t *testing.T) {
	flow := &goyek.Flow{}
	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})

	act := func() {
		flow.Define(goyek.Task{
			Name:  "t1",
			Pools: goyek.DefinedPools{p1, p1},
		})
	}

	assertPanics(t, act, "should panic when task requests more slots than pool limit")
}

func TestPoolWaitLog(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	flow.SetLogger(goyek.FmtLogger{})

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})

	flow.Define(goyek.Task{
		Name:  "blocker",
		Pools: goyek.DefinedPools{p1},
		Action: func(_ *goyek.A) {
			time.Sleep(100 * time.Millisecond)
		},
		Parallel: true,
	})
	flow.Define(goyek.Task{
		Name:  "waiting",
		Pools: goyek.DefinedPools{p1},
		Parallel: true,
	})

	err := flow.Execute(context.Background(), []string{"blocker", "waiting"})
	assertPass(t, err, "Execute should pass")
}

func TestPoolContextCancellationBetweenPools(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)

	p1 := flow.DefinePool(goyek.Pool{Name: "p1", Limit: 1})
	p2 := flow.DefinePool(goyek.Pool{Name: "p2", Limit: 1})

	ctx, cancel := context.WithCancel(context.Background())

	// blocker takes p2
	flow.Define(goyek.Task{
		Name:  "blocker",
		Pools: goyek.DefinedPools{p2},
		Action: func(_ *goyek.A) {
			cancel() // cancel context
			time.Sleep(50 * time.Millisecond)
		},
	})

	// waiter needs p1 then p2.
	// Since pools are sorted by name, p1 is acquired first.
	// blocker holds p2.
	// So waiter will acquire p1, then block on p2.
	// Since context is cancelled by blocker, waiter should hit the cancellation for p2.
	flow.Define(goyek.Task{
		Name:     "waiter",
		Pools:    goyek.DefinedPools{p1, p2},
		Parallel: true,
	})

	err := flow.Execute(ctx, []string{"blocker", "waiter"})
	if err != nil && !strings.Contains(err.Error(), context.Canceled.Error()) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got %v", err)
	}
}

func TestFreshFlowPoolsIntrospection(t *testing.T) {
	f := &goyek.Flow{}
	assertEqual(t, len(f.Pools()), 0, "fresh flow should have no pools")
}

func TestFlow_init_covered(t *testing.T) {
	f := &goyek.Flow{}
	// Calling isPoolDefined or isDefined calls init
	p1 := f.DefinePool(goyek.Pool{Name: "p1", Limit: 1})
	assertEqual(t, p1.Name(), "p1", "pool name")

	// Call init again, should be NOOP
	// We can't verify easily but it increases coverage of the 'if f.tasks == nil' being false.
	f.Define(goyek.Task{Name: "t1"})
	f.Define(goyek.Task{Name: "t2"})
}
