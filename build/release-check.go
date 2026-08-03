package main

import "github.com/goyek/goyek/v3"

var _ = goyek.Define(goyek.Task{
	Name:  "release-check",
	Usage: "checks required before a release",
	Deps: goyek.Deps{
		ci,
		apiDiff,
		govulncheck,
	},
})
