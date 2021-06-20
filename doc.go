/*
Package goyek helps implementing build automation.
It is intended to be used in concert with the "go run" command,
to run a program which implements the build pipeline (called flow).
A flow consists of a set of registered tasks.
A task has a name, can have a defined action, which is a function with signature
	func (*goyek.TF)
and can have dependencies (already defined tasks).

When the flow is executed for given tasks,
then the tasks' actions are run in the order defined by their dependencies.
The task's dependencies are run in a recursive manner, however each is going to be run at most once.

The flow is interrupted in case a action fails.
Within these functions, use the Error, Fail or related methods to signal failure.
*/
package goyek
