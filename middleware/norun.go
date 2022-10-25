package middleware

import "github.com/goyek/goyek/v2"

// NoRun is a middleware which omits running the provided actions.
func NoRun(tasks []string) func(goyek.Runner) goyek.Runner {
	return func(next goyek.Runner) goyek.Runner {
		return func(in goyek.Input) goyek.Result {
			for _, task := range tasks {
				if in.TaskName == task {
					return goyek.Result{}
				}
			}

			return next(in)
		}
	}
}
