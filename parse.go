package goyek

// SplitTasks splits command line arguments into tasks and the rest.
// Tasks are identified as non-flag arguments at the beginning.
// The rest includes flags and any arguments after "--".
//
// This function does not parse flags - it only separates tasks from flags/args.
// To parse flags, use flag.Parse() or similar with the returned rest slice.
//
// Examples:
//   - [task1, task2] -> tasks: [task1, task2], rest: []
//   - [task1, -v] -> tasks: [task1], rest: [-v]
//   - [task1, --, arg1, arg2] -> tasks: [task1], rest: [--, arg1, arg2]
//   - [task1, -v, --, arg1] -> tasks: [task1], rest: [-v, --, arg1]
func SplitTasks(args []string) (tasks, rest []string) {
	flagsStart := -1
	for i, arg := range args {
		// Check if this looks like a flag (starts with -) or separator (--).
		if len(arg) > 0 && arg[0] == '-' {
			flagsStart = i
			break
		}
		// This is a task.
		tasks = append(tasks, arg)
	}
	if flagsStart >= 0 {
		rest = args[flagsStart:]
	}
	return tasks, rest
}
