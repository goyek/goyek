package main

import "github.com/goyek/goyek/v3"

var ci = goyek.Define(goyek.Task{
	Name:  "ci",
	Usage: "CI build pipeline",
	Deps: goyek.Deps{
		all,
		diff,
	},
})
