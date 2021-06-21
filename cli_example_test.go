package goyek_test

import "github.com/goyek/goyek"

func Example() {
	flow := &goyek.Flow{}

	task1 := flow.Register(goyek.Task{
		Name:  "task-1",
		Usage: "Print Go version",
		Action: func(p *goyek.Progress) {
			if err := p.Cmd("go", "version").Run(); err != nil {
				p.Fatal(err)
			}
		},
	})

	task2 := flow.Register(goyek.Task{
		Name: "task-2",
		Action: func(p *goyek.Progress) {
			p.Skip("skipping")
		},
	})

	task3 := flow.Register(goyek.Task{
		Name: "task-3",
		Action: func(p *goyek.Progress) {
			p.Error("hello from", p.Name())
			p.Log("this will be printed")
		},
	})

	flow.Register(goyek.Task{
		Name: "all",
		Deps: goyek.Deps{task1, task2, task3},
	})

	flow.Main()
}
