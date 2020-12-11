package taskflow_test

import "github.com/pellared/taskflow"

func Example() {
	tasks := &taskflow.Taskflow{}

	task1 := tasks.MustRegister(taskflow.Task{
		Name:        "task-1",
		Description: "Print Go version",
		Command:     taskflow.Exec("go", "version"),
	})

	task2 := tasks.MustRegister(taskflow.Task{
		Name: "task-2",
		Command: func(tf *taskflow.TF) {
			tf.Skipf("skipping")
		},
	})

	task3 := tasks.MustRegister(taskflow.Task{
		Name: "task-3",
		Command: func(tf *taskflow.TF) {
			tf.Errorf("hello from " + tf.Name())
			tf.Logf("this will be printed")
		},
	})

	tasks.MustRegister(taskflow.Task{
		Name:         "all",
		Dependencies: taskflow.Deps{task1, task2, task3},
	})

	tasks.Main()
}
