package goyek_test

import "github.com/goyek/goyek"

func Example() {
	flow := &goyek.Flow{}

	task1 := flow.Register(goyek.Task{
		Name:  "task-1",
		Usage: "Print Go version",
		Action: func(a *goyek.A) {
			if err := a.Cmd("go", "version").Run(); err != nil {
				a.Fatal(err)
			}
		},
	})

	task2 := flow.Register(goyek.Task{
		Name: "task-2",
		Action: func(a *goyek.A) {
			a.Skip("skipping")
		},
	})

	task3 := flow.Register(goyek.Task{
		Name: "task-3",
		Action: func(a *goyek.A) {
			a.Error("hello from", a.Name())
			a.Log("this will be printed")
		},
	})

	flow.Register(goyek.Task{
		Name: "all",
		Deps: goyek.Deps{task1, task2, task3},
	})

	flow.Main()
}
