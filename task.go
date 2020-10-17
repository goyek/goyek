package taskflow

type Task struct {
	Name         string
	Description  string
	Command      func(tf *TF)
	Dependencies Deps
}

type Deps []RegisteredTask
