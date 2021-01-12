/*
Package taskflow helps implementing build automation.
It is intended to be used in concert with the "go run" command,
to run a program which implements the build pipeline (called taskflow).
A taskflow consists of a set of registered tasks.
A task has a name, can have a defined command, which is a function with signature
	func (*taskflow.TF)
and can have dependencies (already defined tasks).

When the taskflow is executed for given tasks,
then the tasks' commands are run in the order defined by their dependencies.
The task's dependencies are run in a recusrive manner, however each is going to be run at most once.

The taskflow is interupted in case a command fails.
Within these functions, use the Error, Fail or related methods to signal failure.
*/
package taskflow
