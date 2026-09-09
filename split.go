package goyek

// SplitTasks splits command line arguments into tasks and the rest.
// Tasks are identified as non-flag arguments at the beginning.
// The rest includes flags and any arguments after "--".
//
// This function does not parse flags, it only separates tasks from flags/args.
// To parse flags, you can use [flag.FlagSet.Parse] with the returned rest slice.
// A program that does not accept positional arguments can support the syntax
// "[tasks] [flags]" as follows:
//
//	func main() {
//		tasks, args := goyek.SplitTasks(os.Args[1:])
//		if err := flag.CommandLine.Parse(args); err != nil {
//			fmt.Fprintln(goyek.Output(), err)
//			os.Exit(2)
//		}
//		if flag.NArg() > 0 {
//			fmt.Fprintln(goyek.Output(), "unexpected arguments:", flag.Args())
//			os.Exit(2)
//		}
//		goyek.Main(tasks)
//	}
//
// Programs that intentionally accept positional arguments should require them
// to follow an explicit "--", giving the syntax
// "[tasks] [flags] [--] [args]". Do not accept every value in [flag.Args] as
// intentional: flag parsing stops at the first positional argument, so a task
// placed after a flag would otherwise be mistaken for a task argument.
//
// Examples:
//   - [task1, task2] -> tasks: [task1, task2], rest: nil
//   - [task1, -v] -> tasks: [task1], rest: [-v]
//   - [task1, --, arg1, arg2] -> tasks: [task1], rest: [--, arg1, arg2]
//   - [task1, -v, --, arg1] -> tasks: [task1], rest: [-v, --, arg1]
func SplitTasks(args []string) (tasks, rest []string) {
	flagsStart := -1
	for i, arg := range args {
		// Check if this looks like a flag (starts with -) or separator (--).
		// Single "-" is treated as a non-flag argument.
		if len(arg) > 1 && arg[0] == '-' {
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
