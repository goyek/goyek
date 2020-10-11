package taskflow

// Task TODO.
type Task struct {
	Name         string
	Command      func(*TF)
	Dependencies []Dependency
	// Parallelize bool TODO.
}
