package taskflow_test

import "github.com/pellared/taskflow"

func Example() {
	flow := taskflow.New()

	task1 := flow.MustRegister(taskflow.Task{
		Name:        "task-1",
		Description: "Print Go version",
		Command:     taskflow.Exec("go", "version"),
	})

	task2 := flow.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(tf *taskflow.TF) {
			tf.Skip("skipping")
		},
	})

	task3 := flow.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(tf *taskflow.TF) {
			tf.Error("hello from", tf.Name())
			tf.Log("this will be printed")
		},
	})

	flow.MustRegister(taskflow.Task{
		Name:         "all",
		Dependencies: taskflow.Deps{task1, task2, task3},
	})

	flow.Main()
}
