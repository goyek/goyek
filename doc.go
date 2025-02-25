/*
Package goyek helps implementing task automation.

# Defining tasks

Use [Define] to register a [Task].
Each task consists of an Action function that executes when the task runs.

Tasks can have dependencies, set via Deps. By default, dependencies run sequentially,
but setting Parallel allows a task to run concurrently with other parallel tasks.

A task executes at most once per [Flow.Main] or [Flow.Execute] call.
It is valid to define a task with dependencies but no action.

A default task can be assigned using [SetDefault].

# Running programs

For executing external programs, use [github.com/goyek/x/cmd.Exec],
which covers most cases. See [#60] and [#307] for details on why
his feature is not built-in. In some cases, you may prefer [os/exec].

# Customization

You can customize output and behavior using [SetOutput], [SetLogger], [SetUsage],
and [Execute] (as an alternative to [Main]).

Middlewares can be integrated using [Use] and [UseExecutor]
for additional functionality, such as generating task execution reports,
adding retry logic, exporting execution telemetry.

Basic middlewares are available in [github.com/goyek/goyek/v2/middleware].

For a convenient setup, [github.com/goyek/x/boot.Main] applies commonly
used middlewares and defines configurable flags.
Additional customizations are available in [github.com/goyek/x].

[#60]: https://github.com/goyek/goyek/issues/60
[#307]: https://github.com/goyek/goyek/issues/307
*/
package goyek
