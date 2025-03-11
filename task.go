package goyek

// Task represents a named task that can have action and dependencies.
type Task struct {
	// Name uniquely identifies the task.
	// It cannot be empty and should be easily representable on the CLI.
	Name string

	// Usage provides information what the task does.
	Usage string

	// Action is function that is called when the task is run.
	// A task can have only dependencies and no action to act as a pipeline.
	Action func(a *A)

	// Deps is a collection of defined tasks
	// that need to be run before this task is executed.
	Deps Deps

	// Parallel marks that this task can be run in parallel
	// with (and only with) other parallel tasks.
	Parallel bool
}

// DefinedTask represents a task that has been defined.
// It can be used as a dependency for another task.
type DefinedTask struct {
	task *Task
	flow *Flow
}

// Deps represents a collection of dependencies.
type Deps []*DefinedTask

// Name returns the name of the task.
func (r *DefinedTask) Name() string {
	return r.task.Name
}

// SetName changes the name of the task.
func (r *DefinedTask) SetName(s string) {
	if _, ok := r.flow.tasks[s]; ok {
		panic("task with the same name is already defined")
	}
	oldName := r.task.Name
	snap := r.flow.tasks[oldName]
	snap.Name = s
	r.flow.tasks[s] = snap
	delete(r.flow.tasks, oldName)
}

// Usage returns the description of the task.
func (r *DefinedTask) Usage() string {
	return r.task.Usage
}

// SetUsage sets the description of the task.
func (r *DefinedTask) SetUsage(s string) {
	r.task.Usage = s
}

// Action returns the action of the task.
func (r *DefinedTask) Action() func(a *A) {
	return r.task.Action
}

// SetAction changes the action of the task.
func (r *DefinedTask) SetAction(fn func(a *A)) {
	r.task.Action = fn
}

// Deps returns all task's dependencies.
func (r *DefinedTask) Deps() Deps {
	count := len(r.task.Deps)
	if count == 0 {
		return nil
	}
	deps := make(Deps, count)
	copy(deps, r.task.Deps)
	return deps
}

// SetDeps sets all task's dependencies.
func (r *DefinedTask) SetDeps(deps Deps) {
	count := len(deps)
	if count == 0 {
		r.task.Deps = nil
		return
	}

	for _, dep := range deps {
		if !r.flow.isDefined(dep.Name(), dep.flow) {
			panic("dependency was not defined: " + dep.Name())
		}
	}

	visited := map[string]bool{}
	if ok := r.noCycle(deps, visited); !ok {
		panic("circular dependency")
	}
	depNames := make(Deps, count)
	copy(depNames, deps)
	r.task.Deps = depNames
}

func (r *DefinedTask) noCycle(deps Deps, visited map[string]bool) bool {
	if len(deps) == 0 {
		return true
	}
	for _, dep := range deps {
		name := dep.Name()
		if visited[name] {
			return true
		}
		visited[name] = true
		if name == r.task.Name {
			return false
		}
		if !r.noCycle(dep.Deps(), visited) {
			return false
		}
	}
	return true
}
