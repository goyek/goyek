package middleware

import "github.com/goyek/goyek/v3"

// DryRun is a middleware which omits running the actions.
func DryRun(goyek.Runner) goyek.Runner {
	return func(goyek.Input) goyek.Result {
		return goyek.Result{}
	}
}
