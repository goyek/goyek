package goyek_test

import "github.com/goyek/goyek"

func Example() {
	flow := &goyek.Taskflow{}

	task1 := flow.Register(goyek.Task{
		Name:  "task-1",
		Usage: "Print Go version",
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("go", "version").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	})

	task2 := flow.Register(goyek.Task{
		Name: "task-2",
		Action: func(tf *goyek.TF) {
			tf.Skip("skipping")
		},
	})

	task3 := flow.Register(goyek.Task{
		Name: "task-3",
		Action: func(tf *goyek.TF) {
			tf.Error("hello from", tf.Name())
			tf.Log("this will be printed")
		},
	})

	flow.Register(goyek.Task{
		Name: "all",
		Deps: goyek.Deps{task1, task2, task3},
	})

	flow.Main()
}
