package goyek

import (
	"flag"
	"os"
)

// Parse parses command line arguments according to the syntax:
// [tasks] [flags] [--] [args]
//
// Tasks are identified as non-flag arguments at the beginning.
// Flags are parsed using the provided FlagSet.
// Everything after "--" becomes positional args available via flagSet.Args().
//
// If flagSet is nil, flag.CommandLine is used.
// If args is nil, os.Args[1:] is used.
//
// Examples:
//   - "task1 task2" -> tasks: [task1, task2], flagSet.Args(): []
//   - "task1 -v" -> tasks: [task1], flag -v is parsed, flagSet.Args(): []
//   - "task1 -- arg1 arg2" -> tasks: [task1], flagSet.Args(): [arg1, arg2]
//   - "task1 -v -- arg1" -> tasks: [task1], flag -v is parsed, flagSet.Args(): [arg1]
func Parse(args []string, flagSet *flag.FlagSet) ([]string, error) {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	if args == nil {
		args = os.Args[1:]
	}

	// Find the separator "--" if it exists.
	separatorIdx := -1
	for i, arg := range args {
		if arg == "--" {
			separatorIdx = i
			break
		}
	}

	var beforeSeparator []string
	var afterSeparator []string

	if separatorIdx >= 0 {
		beforeSeparator = args[:separatorIdx]
		afterSeparator = args[separatorIdx+1:]
	} else {
		beforeSeparator = args
	}

	// Extract tasks (non-flag arguments at the beginning).
	tasks := []string{}
	flagsStart := -1

	for i, arg := range beforeSeparator {
		// Check if this looks like a flag (starts with -).
		if len(arg) > 0 && arg[0] == '-' {
			flagsStart = i
			break
		}
		// This is a task.
		tasks = append(tasks, arg)
	}

	// Build args for flag.Parse()
	// If we have args after separator, append them after flags.
	flagArgs := []string{}
	if flagsStart >= 0 {
		flagArgs = beforeSeparator[flagsStart:]
	}

	// Append the positional args after "--" separator.
	if len(afterSeparator) > 0 {
		flagArgs = append(flagArgs, afterSeparator...)
	}

	if err := flagSet.Parse(flagArgs); err != nil {
		return nil, err
	}
	return tasks, nil
}
