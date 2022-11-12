package goyek

// Task represents a named task that can be defined.
// It can consist of a action (function that will be called when task is run),
// dependencies (tasks which has to be run before this one).
type Task struct {
	// Name uniquely identifies the task.
	// It cannot be empty and should be easily representable on the CLI.
	Name string

	// Usage provides information what the task does.
	Usage string

	// Action executes the task in the given flow context.
	// A task can be defined without a action to act as a "collector" task for a list of dependencies.
	Action func(a *A)

	// Deps is a collection of defined tasks that need to be run before this task is executed.
	Deps Deps
}

// DefinedTask represents a task that has been defined.
// It can be used as a dependency for another task.
type DefinedTask struct {
	*taskSnapshot
	flow *Flow
}

// Deps represents a collection of dependencies.
type Deps []*DefinedTask

// Name returns the name of the task.
func (r *DefinedTask) Name() string {
	return r.name
}

// SetName changes the name of the task.
func (r *DefinedTask) SetName(s string) {
	if _, ok := r.flow.tasks[s]; ok {
		panic("task with the same name is already defined")
	}
	oldName := r.name
	snap := r.flow.tasks[oldName]
	snap.name = s
	r.flow.tasks[s] = snap
	delete(r.flow.tasks, oldName)
}

// Usage returns the description of the task.
func (r *DefinedTask) Usage() string {
	return r.usage
}

// SetUsage sets the description of the task.
func (r *DefinedTask) SetUsage(s string) {
	r.usage = s
}

// Action returns the action of the task.
func (r *DefinedTask) Action() func(a *A) {
	return r.action
}

// SetAction changes the action of the task.
func (r *DefinedTask) SetAction(fn func(a *A)) {
	r.action = fn
}

// Deps returns all task's dependencies.
func (r *DefinedTask) Deps() Deps {
	count := len(r.deps)
	if count == 0 {
		return nil
	}
	deps := make(Deps, 0, count)
	for _, dep := range r.deps {
		deps = append(deps, &DefinedTask{r.flow.tasks[dep.name], r.flow})
	}
	return deps
}

// SetDeps sets all task's dependencies.
func (r *DefinedTask) SetDeps(deps Deps) {
	count := len(deps)
	if count == 0 {
		r.deps = nil
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
	depNames := make([]*taskSnapshot, 0, count)
	for _, dep := range deps {
		depNames = append(depNames, dep.taskSnapshot)
	}

	r.deps = depNames
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
		if name == r.name {
			return false
		}
		if !r.noCycle(dep.Deps(), visited) {
			return false
		}
	}
	return true
}
