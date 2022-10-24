package goyek_test

import (
	"context"
	"io/ioutil"
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
	err := flow.Execute(context.Background(), "new")
	assertPass(t, err, "should pass")
	assertTrue(t, called, "should call the action")
}

func TestDefinedTask_SetDeps(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	called := false
	t1 := flow.Define(goyek.Task{Name: "one", Action: func(tf *goyek.TF) { called = true }})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})
	t3 := flow.Define(goyek.Task{Name: "three"})

	t3.SetDeps(goyek.Deps{t2})

	got := t3.Deps()
	assertEqual(t, got, goyek.Deps{t2}, "should update the dependencies")

	err := flow.Execute(context.Background(), "three")
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

	var want goyek.Deps
	got := t2.Deps()
	assertEqual(t, got, want, "should clear the dependencies")

	err := flow.Execute(context.Background(), "two")
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
