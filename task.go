package taskflow

// Task TODO.
type Task struct {
	Name         string
	Command      func(*TF) error
	Dependencies []Dependency
}
