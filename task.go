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
	name     string
	usage    string
	deps     []*DefinedTask
	action   func(a *A)
	parallel bool
	flow     *Flow
}

// Deps represents a collection of dependencies.
type Deps []*DefinedTask

// Name returns the name of the task.
func (r *DefinedTask) Name() string {
	r.flow.mu.RLock()
	defer r.flow.mu.RUnlock()
	return r.name
}

// SetName changes the name of the task.
func (r *DefinedTask) SetName(s string) {
	r.flow.mu.Lock()
	defer r.flow.mu.Unlock()
	if _, ok := r.flow.tasks[s]; ok {
		panic("task with the same name is already defined")
	}
	oldName := r.name
	r.flow.tasks[s] = r
	delete(r.flow.tasks, oldName)
	r.name = s
}

// Usage returns the description of the task.
func (r *DefinedTask) Usage() string {
	r.flow.mu.RLock()
	defer r.flow.mu.RUnlock()
	return r.usage
}

// SetUsage sets the description of the task.
func (r *DefinedTask) SetUsage(s string) {
	r.flow.mu.Lock()
	defer r.flow.mu.Unlock()
	r.usage = s
}

// Action returns the action of the task.
func (r *DefinedTask) Action() func(a *A) {
	r.flow.mu.RLock()
	defer r.flow.mu.RUnlock()
	return r.action
}

// SetAction changes the action of the task.
func (r *DefinedTask) SetAction(fn func(a *A)) {
	r.flow.mu.Lock()
	defer r.flow.mu.Unlock()
	r.action = fn
}

// Deps returns all task's dependencies.
func (r *DefinedTask) Deps() Deps {
	r.flow.mu.RLock()
	defer r.flow.mu.RUnlock()
	if len(r.deps) == 0 {
		return nil
	}
	deps := make(Deps, len(r.deps))
	copy(deps, r.deps)
	return deps
}

// SetDeps sets all task's dependencies.
func (r *DefinedTask) SetDeps(deps Deps) {
	r.flow.mu.Lock()
	defer r.flow.mu.Unlock()
	if len(deps) == 0 {
		r.deps = nil
		return
	}

	for _, dep := range deps {
		if !r.flow.isDefinedLocked(dep.name, dep.flow) {
			panic("dependency was not defined: " + dep.name)
		}
	}

	visited := map[string]bool{}
	if ok := r.noCycle(deps, visited); !ok {
		panic("circular dependency")
	}
	r.deps = deps
}

func (r *DefinedTask) noCycle(deps Deps, visited map[string]bool) bool {
	if len(deps) == 0 {
		return true
	}
	for _, dep := range deps {
		name := dep.name
		if visited[name] {
			continue // already checked this branch
		}
		visited[name] = true
		if name == r.name {
			return false
		}
		if !r.noCycle(dep.deps, visited) {
			return false
		}
	}
	return true
}
