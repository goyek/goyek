package main

import "github.com/goyek/goyek/v2"

var all = flow.Define(goyek.Task{
	Name:  "all",
	Usage: "build pipeline",
	Deps: goyek.Deps{
		mod,
		install,
		lint,
		test,
	},
})
