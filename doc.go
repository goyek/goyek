/*
Package goyek helps implementing task automation.

# Defining tasks

Use [Define] to register a [Task].

Action is a function which executes when a task is running.

You can add dependencies to task by setting Deps.

By default, the dependencies are running in sequential order.
Parallel can be set to allow a task to be run in parallel with other parallel tasks.

Each task during [Flow.Main] or [Flow.Execute] runs at most once.

It is valid to have a task with dependencies and no action.

A default task can be assigned using [SetDefault].

# Running programs

You can use the [github.com/goyek/x/cmd.Exec] that should cover most use cases.

[#60] and [#307] explain why this feature is not out-of-the-box.
In some cases, you may prefer to use [os/exec].

# Customizing

You can customize the default output by using
[SetOutput], [SetLogger], [SetUsage], [Execute] (instead of [Main]).

[Use] and [UseExecutor] are provided to allow using different types of middlewares.
You can use a middleware, for example to: generate a task execution report,
add retry logic, export task execution telemetry, etc.
Some basic middlewares are provided in [github.com/goyek/goyek/v2/middleware] package.

[github.com/goyek/x/boot.Main] convenient function sets the most commonly used middlewares
nd defines flags to configure them.

Some reusable customization are offered by [github.com/goyek/x].

[#60]: https://github.com/goyek/goyek/issues/60
[#307]: https://github.com/goyek/goyek/issues/307
*/
package goyek
