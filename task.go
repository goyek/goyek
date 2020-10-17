package taskflow

// Task TODO.
type Task struct {
	Name         string
	Description  string
	Command      func(*TF)
	Dependencies Deps
}

// Deps TODO.
type Deps []RegisteredTask
