package middleware

import "github.com/goyek/goyek/v2"

// DryRun is a middleware which omits running the actions.
func DryRun(goyek.Runner) goyek.Runner {
	return func(in goyek.Input) goyek.Result {
		return goyek.Result{}
	}
}
