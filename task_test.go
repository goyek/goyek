package goyek_test

import (
	"context"
	"io"
	"reflect"
	"runtime"
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestDefinedTask_comparable(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "task"})
	sameTask := flow.Tasks()[0]

	tasks := map[goyek.DefinedTask]bool{*task: true}

	assertTrue(t, tasks[*sameTask], "should remain usable as a map key")
}

func TestDefinedTask_copyRemainsHandle(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "task"})
	taskCopy := *task

	taskCopy.SetUsage("updated through copy")

	assertEqual(t, task.Usage(), "updated through copy", "should share the task state")
}

func TestDefinedTask_SetName(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(*goyek.A) { called = true }})

	task.SetName("new")

	got := task.Name()
	assertEqual(t, got, "new", "should update the name")
	err := flow.Execute(context.Background(), []string{"new"})
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call the action")
}

func TestDefinedTask_SetName_for_default(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(*goyek.A) { called = true }})
	flow.SetDefault(task)

	task.SetName("new")

	got := task.Name()
	assertEqual(t, got, "new", "should update the name")
	err := flow.Execute(context.Background(), nil)
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call the action")
}

func TestDefinedTask_SetName_for_depenency(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(*goyek.A) { called = true }})
	flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{task}})

	task.SetName("new")

	err := flow.Execute(context.Background(), []string{"two"})
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call the dependency with changed name")
}

func TestDefinedTask_SetName_conflict(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	task := flow.Define(goyek.Task{Name: "one"})
	flow.Define(goyek.Task{Name: "two"})

	act := func() { task.SetName("two") }

	assertPanics(t, act, "should not allow setting existing task name")
}

func TestDefinedTask_SetUsage(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	task := flow.Define(goyek.Task{Name: "one"})

	task.SetUsage("good task")
	got := flow.Tasks()[0].Usage()

	assertEqual(t, got, "good task", "should update the usage")
}

func TestDefinedTask_SetAction(t *testing.T) {
	getFuncName := func(fn func(a *goyek.A)) string {
		return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	}

	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	originalNotCalled := true
	task := flow.Define(goyek.Task{Name: "one", Action: func(*goyek.A) { originalNotCalled = false }})

	newCalled := false
	fn := func(*goyek.A) { newCalled = true }
	task.SetAction(fn)
	want := getFuncName(fn)
	got := getFuncName(task.Action())

	assertEqual(t, got, want, "should update the action")
	err := flow.Execute(context.Background(), []string{"one"})
	assertPass(t, err, "should pass")
	assertTrue(t, originalNotCalled, "should not call the previous action")
	assertTrue(t, newCalled, "should not call the new action")
}

func TestDefinedTask_SetDeps(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	called := false
	t1 := flow.Define(goyek.Task{Name: "one", Action: func(*goyek.A) { called = true }})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})
	t3 := flow.Define(goyek.Task{Name: "three"})

	t3.SetDeps(goyek.Deps{t1, t2})

	got := t3.Deps()
	assertEqual(t, got, goyek.Deps{t1, t2}, "should update the dependencies")

	err := flow.Execute(context.Background(), []string{"three"})
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call transitive dependency of t3")
}

func TestDefinedTask_SetDeps_copiesInput(t *testing.T) {
	flow := &goyek.Flow{}
	root := flow.Define(goyek.Task{Name: "root"})
	task := flow.Define(goyek.Task{Name: "task"})
	deps := goyek.Deps{root}
	task.SetDeps(deps)
	dependent := flow.Define(goyek.Task{Name: "dependent", Deps: goyek.Deps{task}})

	// This would create a cycle if SetDeps retained the caller's slice.
	deps[0] = dependent

	got := task.Deps()
	requireEqual(t, len(got), 1, "should keep one dependency")
	assertEqual(t, got[0], root, "should retain the validated dependency")
}

func TestDefinedTask_SetDeps_clear(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	notCalled := true
	t1 := flow.Define(goyek.Task{Name: "one", Action: func(*goyek.A) { notCalled = false }})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})

	t2.SetDeps(nil)

	got := t2.Deps()
	var noDeps goyek.Deps
	assertEqual(t, got, noDeps, "should clear the dependencies")

	err := flow.Execute(context.Background(), []string{"two"})
	assertPass(t, err, "should pass")
	assertTrue(t, notCalled, "should not call any dependency")
}

func TestDefinedTask_SetDeps_circular(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(io.Discard)
	t1 := flow.Define(goyek.Task{Name: "one"})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})
	t3 := flow.Define(goyek.Task{Name: "three", Deps: goyek.Deps{t2}})

	act := func() {
		t1.SetDeps(goyek.Deps{t3})
	}

	assertPanics(t, act, "should panic in case of a cyclic dependency")
}

func TestDefinedTask_SetDeps_dep(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "my-task"})
	otherFlow := &goyek.Flow{}
	otherTask := otherFlow.Define(goyek.Task{Name: "different-flow"})

	act := func() { task.SetDeps(goyek.Deps{otherTask}) }

	assertPanics(t, act, "should not be possible use dependencies from different flow")
}

func TestDefinedTask_SetDeps_staleDependency(t *testing.T) {
	flow := &goyek.Flow{}
	stale := flow.Define(goyek.Task{Name: "task"})
	flow.Undefine(stale)
	flow.Define(goyek.Task{Name: "task"})
	consumer := flow.Define(goyek.Task{Name: "consumer"})

	act := func() { consumer.SetDeps(goyek.Deps{stale}) }

	assertPanics(t, act, "should reject a stale dependency")
	assertEqual(t, consumer.Deps(), goyek.Deps(nil), "should not update dependencies")
}

func TestDefinedTask_mutatorsRejectStaleHandle(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*goyek.DefinedTask, *goyek.DefinedTask)
	}{
		{
			name: "name",
			mutate: func(stale, _ *goyek.DefinedTask) {
				stale.SetName("renamed")
			},
		},
		{
			name: "usage",
			mutate: func(stale, _ *goyek.DefinedTask) {
				stale.SetUsage("stale usage")
			},
		},
		{
			name: "action",
			mutate: func(stale, _ *goyek.DefinedTask) {
				stale.SetAction(func(*goyek.A) {})
			},
		},
		{
			name: "dependencies",
			mutate: func(stale, _ *goyek.DefinedTask) {
				stale.SetDeps(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flow := &goyek.Flow{}
			stale := flow.Define(goyek.Task{Name: "task"})
			flow.Undefine(stale)
			replacement := flow.Define(goyek.Task{Name: "task", Usage: "replacement"})

			act := func() { tt.mutate(stale, replacement) }

			assertPanics(t, act, "should reject a stale task handle")
			got := flow.Tasks()
			requireEqual(t, len(got), 1, "should retain the replacement task")
			assertEqual(t, got[0], replacement, "should not replace the current task")
			assertEqual(t, got[0].Usage(), "replacement", "should not reconfigure the current task")
		})
	}
}
