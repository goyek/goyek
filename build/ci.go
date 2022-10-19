package main

import "github.com/goyek/goyek/v2"

var _ = goyek.Define(goyek.Task{
	Name:  "ci",
	Usage: "CI build pipeline",
	Deps: goyek.Deps{
		all,
		diff,
	},
})
