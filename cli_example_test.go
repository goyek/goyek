package taskflow_test

import "github.com/pellared/taskflow"

func Example() {
	flow := taskflow.New()

	task1 := flow.Register(taskflow.Task{
		Name:    "task-1",
		Usage:   "Print Go version",
		Command: taskflow.Exec("go", "version"),
	})

	task2 := flow.Register(taskflow.Task{
		Name: "task-2",
		Command: func(tf *taskflow.TF) {
			tf.Skip("skipping")
		},
	})

	task3 := flow.Register(taskflow.Task{
		Name: "task-3",
		Command: func(tf *taskflow.TF) {
			tf.Error("hello from", tf.Name())
			tf.Log("this will be printed")
		},
	})

	flow.Register(taskflow.Task{
		Name: "all",
		Deps: taskflow.Deps{task1, task2, task3},
	})

	flow.Main()
}
