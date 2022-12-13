package goyek_test

import (
	"context"
	"io/ioutil"
	"reflect"
	"runtime"
	"testing"

	"github.com/goyek/goyek/v2"
)

func TestDefinedTaskSetName(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(a *goyek.A) { called = true }})

	want := "new"
	task.SetName(want)

	if got := task.Name(); got != want {
		t.Errorf("should update the name\ngot = %q\nwant = %q", got, want)
	}
	if err := flow.Execute(context.Background(), []string{want}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if !called {
		t.Errorf("should call the action")
	}
}

func TestDefinedTaskSetNameForDefault(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(a *goyek.A) { called = true }})
	flow.SetDefault(task)

	want := "new"
	task.SetName(want)

	if got := task.Name(); got != want {
		t.Errorf("should update the name\ngot = %q\nwant = %q", got, want)
	}
	if err := flow.Execute(context.Background(), nil); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if !called {
		t.Errorf("should call the action")
	}
}

func TestDefinedTaskSetNameForDep(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	called := false
	task := flow.Define(goyek.Task{Name: "one", Action: func(a *goyek.A) { called = true }})
	flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{task}})

	task.SetName("new")

	if err := flow.Execute(context.Background(), []string{"two"}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if !called {
		t.Errorf("should call the dependency with changed name")
	}
}

func TestDefinedTaskSetNameConflict(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	task := flow.Define(goyek.Task{Name: "one"})
	flow.Define(goyek.Task{Name: "two"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("should not allow setting existing task name")
		}
	}()
	task.SetName("two")
}

func TestDefinedTaskSetUsage(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	task := flow.Define(goyek.Task{Name: "one"})

	want := "good task"
	task.SetUsage(want)

	if got := flow.Tasks()[0].Usage(); got != want {
		t.Errorf("should update the usage\ngot = %q\nwant = %q", got, want)
	}
}

func TestDefinedTaskSetAction(t *testing.T) {
	getFuncName := func(fn func(a *goyek.A)) string {
		return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	}

	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var originalCalled, newCalled bool
	task := flow.Define(goyek.Task{Name: "one", Action: func(a *goyek.A) { originalCalled = true }})

	fn := func(a *goyek.A) { newCalled = true }
	task.SetAction(fn)

	if got, want := getFuncName(task.Action()), getFuncName(fn); got != want {
		t.Errorf("should update the action\ngot = %q\nwant = %q", got, want)
	}
	if err := flow.Execute(context.Background(), []string{"one"}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if originalCalled {
		t.Errorf("should not call the previous action")
	}
	if !newCalled {
		t.Errorf("should not call the new action")
	}
}

func TestDefinedTaskSetDeps(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	called := false
	t1 := flow.Define(goyek.Task{Name: "one", Action: func(a *goyek.A) { called = true }})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})
	t3 := flow.Define(goyek.Task{Name: "three"})

	t3.SetDeps(goyek.Deps{t1, t2})

	if got, want := t3.Deps(), (goyek.Deps{t1, t2}); !reflect.DeepEqual(got, want) {
		t.Errorf("should update the dependencies\ngot = %#v\nwant = %#v", got, want)
	}
	if err := flow.Execute(context.Background(), []string{"three"}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if !called {
		t.Errorf("should call transitive dependency of t3")
	}
}

func TestDefinedTaskSetDepsClear(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	var called bool
	t1 := flow.Define(goyek.Task{Name: "one", Action: func(a *goyek.A) { called = true }})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})

	t2.SetDeps(nil)

	if got := t2.Deps(); got != nil {
		t.Errorf("should clear the dependencies\ngot = %#v\n", got)
	}
	if err := flow.Execute(context.Background(), []string{"two"}); err != nil {
		t.Errorf("should pass, but was: %v", err)
	}
	if called {
		t.Errorf("should not call any dependency")
	}
}

func TestDefinedTaskSetDepsCircular(t *testing.T) {
	flow := &goyek.Flow{}
	flow.SetOutput(ioutil.Discard)
	t1 := flow.Define(goyek.Task{Name: "one"})
	t2 := flow.Define(goyek.Task{Name: "two", Deps: goyek.Deps{t1}})
	t3 := flow.Define(goyek.Task{Name: "three", Deps: goyek.Deps{t2}})

	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic in case of a cyclic dependency")
		}
	}()
	t1.SetDeps(goyek.Deps{t3})
}

func TestDefinedTaskSetDepsBadDep(t *testing.T) {
	flow := &goyek.Flow{}
	task := flow.Define(goyek.Task{Name: "my-task"})
	otherFlow := &goyek.Flow{}
	otherTask := otherFlow.Define(goyek.Task{Name: "different-flow"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("should not be possible use dependencies from different flow")
		}
	}()
	task.SetDeps(goyek.Deps{otherTask})
}
