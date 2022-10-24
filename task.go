package goyek

// Task represents a named task that can be registered.
// It can consist of a action (function that will be called when task is run),
// dependencies (tasks which has to be run before this one)
// parameters (which can be used within the action).
type Task struct {
	// Name uniquely identifies the task.
	// Names cannot  be empty and should be easily representable on the CLI.
	Name string

	// Usage provides information what the task does.
	Usage string

	// Action executes the task in the given flow context.
	// A task can be registered without a action and can act as a "collector" task
	// for a list of dependencies (also called "pipeline").
	Action func(tf *TF)

	// Deps is a collection of registered tasks that need to be run before this task is executed.
	Deps Deps
}

// DefinedTask represents a task that has been registered.
// It can be used as a dependency for another task.
type DefinedTask interface {
	Name() string
	SetName(string)
	Usage() string
	SetUsage(string)
	Action() func(tf *TF)
	SetAction(func(tf *TF))
	Deps() Deps
	SetDeps(Deps)
	snapshot() *taskSnapshot
}

// Deps represents a collection of dependencies.
type Deps []DefinedTask

// registeredTask implements (and encapsulates) DefinedTask.
type registeredTask struct {
	*taskSnapshot
	flow *Flow
}

// Name returns the name of the task.
func (r registeredTask) Name() string {
	return r.name
}

// Name changes the name of the task.
func (r registeredTask) SetName(s string) {
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
func (r registeredTask) Usage() string {
	return r.usage
}

// SetUsage sets the description of the task.
func (r registeredTask) SetUsage(s string) {
	r.usage = s
}

// Action returns the action of the task.
func (r registeredTask) Action() func(tf *TF) {
	return r.action
}

// SetAction changes the action of the task.
func (r registeredTask) SetAction(fn func(tf *TF)) {
	r.action = fn
}

// Deps returns all task's dependencies.
func (r registeredTask) Deps() Deps {
	count := len(r.deps)
	if count == 0 {
		return nil
	}
	deps := make(Deps, 0, count)
	for _, dep := range r.deps {
		deps = append(deps, registeredTask{r.flow.tasks[dep.name], r.flow})
	}
	return deps
}

// Deps returns all task's dependencies.
func (r registeredTask) SetDeps(deps Deps) {
	count := len(deps)
	if count == 0 {
		r.deps = nil
		return
	}

	for _, dep := range deps {
		if !r.flow.isDefined(dep.Name()) {
			panic("dependency was not defined: " + dep.Name())
		}
	}

	visited := map[string]bool{}
	if ok := r.noCycle(deps, visited); !ok {
		panic("circular dependency")
	}
	depNames := make([]*taskSnapshot, 0, count)
	for _, dep := range deps {
		depNames = append(depNames, dep.snapshot())
	}

	r.deps = depNames
}

func (r registeredTask) noCycle(deps Deps, visited map[string]bool) bool {
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

func (r registeredTask) snapshot() *taskSnapshot {
	return r.taskSnapshot
}
