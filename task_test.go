package goyek_test

import (
	"context"
	"io/ioutil"
	"reflect"
	"runtime"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestDefinedTask_SetName(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(tf *goyek.TF) { called = true }})

	task.SetName("new")

	got := task.Name()
	assertEqual(t, got, "new", "should update the name")
	err := flow.Execute(context.Background(), []string{"new"})
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call the action")
}

func TestDefinedTask_SetName_for_default(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(tf *goyek.TF) { called = true }})
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
	flow.SetOutput(ioutil.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(tf *goyek.TF) { called = true }})
	flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{task}})

	task.SetName("new")

	err := flow.Execute(context.Background(), []string{"two"})
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call the dependency with changed name")
}

func TestDefinedTask_SetName_conflict(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	task := flow.Define(goyek.Task{Name: "one"})
	flow.Define(goyek.Task{Name: "two"})

	act := func() { task.SetName("two") }

	assertPanics(t, act, "should not allow setting existing task name")
}

func TestDefinedTask_SetUsage(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	task := flow.Define(goyek.Task{Name: "one"})

	task.SetUsage("good task")
	got := flow.Tasks()[0].Usage()

	assertEqual(t, got, "good task", "should update the usage")
}

func TestDefinedTask_SetAction(t *testing.T) {
	getFuncName := func(fn func(tf *goyek.TF)) string {
		return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	}

	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	originalNotCalled := true
	task := flow.Define(goyek.Task{Name: "one", Action: func(tf *goyek.TF) { originalNotCalled = false }})

	newCalled := false
	fn := func(tf *goyek.TF) { newCalled = true }
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
	flow.SetOutput(ioutil.Discard)
	called := false
	t1 := flow.Define(goyek.Task{Name: "one", Action: func(tf *goyek.TF) { called = true }})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})
	t3 := flow.Define(goyek.Task{Name: "three"})

	t3.SetDeps(goyek.Deps{t1, t2})

	got := t3.Deps()
	assertEqual(t, got, goyek.Deps{t1, t2}, "should update the dependencies")

	err := flow.Execute(context.Background(), []string{"three"})
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call transitive dependency of t3")
}

func TestDefinedTask_SetDeps_clear(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	notCalled := true
	t1 := flow.Define(goyek.Task{Name: "one", Action: func(tf *goyek.TF) { notCalled = false }})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})

	t2.SetDeps(nil)

	got := t2.Deps()
	assertEqual(t, got, noDeps, "should clear the dependencies")

	err := flow.Execute(context.Background(), []string{"two"})
	assertPass(t, err, "should pass")
	assertTrue(t, notCalled, "should not call any dependency")
}

func TestDefinedTask_SetDeps_circular(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
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
