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
//
// A DefinedTask is not safe for concurrent use.
type DefinedTask struct {
	*taskSnapshot
	flow *Flow
}

// Deps represents a collection of dependencies.
type Deps []*DefinedTask

// taskSnapshot contains the mutable state owned by a Flow.
type taskSnapshot struct {
	name     string
	usage    string
	deps     []*taskSnapshot
	action   func(a *A)
	parallel bool
}

// Name returns the name of the task.
func (r *DefinedTask) Name() string {
	return r.name
}

// SetName changes the name of the task.
func (r *DefinedTask) SetName(s string) {
	r.mustBeDefined()
	if _, ok := r.flow.tasks[s]; ok {
		panic("task with the same name is already defined")
	}
	oldName := r.name
	r.flow.tasks[s] = r.taskSnapshot
	delete(r.flow.tasks, oldName)
	r.name = s
}

// Usage returns the description of the task.
func (r *DefinedTask) Usage() string {
	return r.usage
}

// SetUsage sets the description of the task.
func (r *DefinedTask) SetUsage(s string) {
	r.mustBeDefined()
	r.usage = s
}

// Action returns the action of the task.
func (r *DefinedTask) Action() func(a *A) {
	return r.action
}

// SetAction changes the action of the task.
func (r *DefinedTask) SetAction(fn func(a *A)) {
	r.mustBeDefined()
	r.action = fn
}

// Deps returns all task's dependencies.
func (r *DefinedTask) Deps() Deps {
	if len(r.deps) == 0 {
		return nil
	}
	deps := make(Deps, len(r.deps))
	for i, dep := range r.deps {
		deps[i] = &DefinedTask{taskSnapshot: dep, flow: r.flow}
	}
	return deps
}

// SetDeps sets all task's dependencies.
func (r *DefinedTask) SetDeps(deps Deps) {
	r.mustBeDefined()
	snapshots := r.flow.snapshotDeps(deps)

	visited := map[*taskSnapshot]bool{}
	if ok := r.noCycle(snapshots, visited); !ok {
		panic("circular dependency")
	}
	r.deps = snapshots
}

func (r *DefinedTask) noCycle(deps []*taskSnapshot, visited map[*taskSnapshot]bool) bool {
	if len(deps) == 0 {
		return true
	}
	for _, dep := range deps {
		if visited[dep] {
			continue // already checked this branch
		}
		visited[dep] = true
		if dep == r.taskSnapshot {
			return false
		}
		if !r.noCycle(dep.deps, visited) {
			return false
		}
	}
	return true
}

func (r *DefinedTask) mustBeDefined() {
	if !r.flow.isDefined(r) {
		panic("task was not defined: " + r.name)
	}
}
